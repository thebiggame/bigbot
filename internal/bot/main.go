package bot

import (
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/thebiggame/bigbot/internal/avbridge"
	"github.com/thebiggame/bigbot/internal/config"
	"github.com/thebiggame/bigbot/internal/helpers"
	log "github.com/thebiggame/bigbot/internal/log"
	"github.com/thebiggame/bigbot/internal/teamroles"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"reflect"
	"syscall"
)

type BotModule interface {
	Start(context context.Context) error
	DiscordCommands() ([]*discordgo.ApplicationCommand, error)
	HandleDiscordCommand(session *discordgo.Session, interaction *discordgo.InteractionCreate) (handled bool, err error)
}

type BigBot struct {
	DiscordSession *discordgo.Session
	commands       []*discordgo.ApplicationCommand
	modules        []BotModule
}

func New() (*BigBot, error) {
	// get base discord session
	DiscordSession, err := discordgo.New("Bot " + config.RuntimeConfig.Discord.Token)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}

	// create primary bot object and return it
	bot := &BigBot{
		DiscordSession: DiscordSession,
	}
	return bot, nil
}

// WithWANModules loads the modules that should usually run in a cloud / WAN environment.
// These modules should not normally need network access to a LAN event.
func (b *BigBot) WithWANModules() *BigBot {
	// teamRoles
	module := teamroles.New(b.DiscordSession)
	b.modules = append(b.modules, module)
	return b
}

// WithLANModules loads the modules that are relevant to the bot running in a LAN environment,
// e.g. those that require intranet access to function properly.
func (b *BigBot) WithLANModules() *BigBot {
	// avbridge
	module, err := avbridge.New(b.DiscordSession)
	if err != nil {
		panic(err)
	}
	b.modules = append(b.modules, module)
	return b
}

func (b *BigBot) handleDiscordCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	g := new(errgroup.Group)
	for _, m := range b.modules {
		g.Go(func() error {
			handled, err := m.HandleDiscordCommand(s, i)
			if handled {
				log.Debugf("Module %s handled command %v", reflect.TypeOf(m).Elem().Name(), i.ApplicationCommandData().Name)
			}
			return err
		})
	}
	if err := g.Wait(); err != nil {
		// Error occurred.
		log.Error(err)
		log.Debugf("Error occurred while processing: %s", i.Interaction.Data)
		// Figure out how to report it.
		var content string
		if IsCrew, err := helpers.UserIsCrew(s, i.GuildID, i.Member.User); err != nil && IsCrew {
			content = fmt.Sprintf("🚫 **An error occurred while processing your command:**\n```%s```", err)
		} else {
			content = "🚫 **An error occurred while processing your command. Please contact a member of theBIGGAME Crew.**"
		}
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			log.Errorf("Error returning log to client for slash command: %v", err)
		}
	}

}

func (b *BigBot) registerCommands() (err error) {
	var modCmds []*discordgo.ApplicationCommand
	// Collate all slash commands.
	for _, v := range b.modules {
		mC, err := v.DiscordCommands()
		if err != nil {
			return err
		}
		for _, cmd := range mC {
			modCmds = append(modCmds, cmd)
		}
	}
	// Write them out en masse to the guild.
	cmds, err := b.DiscordSession.ApplicationCommandBulkOverwrite(b.DiscordSession.State.User.ID, config.RuntimeConfig.Discord.GuildID, modCmds)
	if err != nil {
		// This shouldn't happen. Bail out
		return fmt.Errorf("error creating commands: %w", err)
	}
	// Write them to our understanding of the commands.
	// We write here with the new command knowledge (from the ApplicationCommandBulkOverwrite) so that we can interact with them later (they'll have IDs).
	copy(b.commands, cmds)
	return nil
}

// TeardownCommands destroys all slash commands on the server associated with this run of the bot.
func (b *BigBot) TeardownCommands() error {
	for _, cmd := range b.commands {
		err := b.DiscordSession.ApplicationCommandDelete(b.DiscordSession.State.User.ID, config.RuntimeConfig.Discord.GuildID, cmd.ID)
		if err != nil {
			return fmt.Errorf("error removing command %s: %w", cmd.Name, err)
		}
		log.Debugf("Removed command %s", cmd.Name)
	}
	log.Info("Commands have been removed successfully.")
	return nil
}

func (b *BigBot) Run() (err error) {
	if config.RuntimeConfig.Discord.Token == "" {
		return errors.New("no discord token provided")
	}

	b.DiscordSession.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Infof("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	err = b.DiscordSession.Open()
	if err != nil {
		return fmt.Errorf("error opening discord connection: %w", err)
	}
	defer b.DiscordSession.Close()

	// Register all module slash commands.
	err = b.registerCommands()
	if err != nil {
		return fmt.Errorf("error registering commands: %w", err)
	}

	// Create app context (this is passed to modules).
	// The signal.NotifyContext is a special context that gets torn down when an interrupt / SIGTERM is received.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// This context is the one that we actually pass - if an initialisation error happens with a module,
	// it propagates out to the rest.
	g, gCtx := errgroup.WithContext(ctx)

	for _, module := range b.modules {
		g.Go(func() error {
			err := module.Start(gCtx)
			if err != nil && !errors.Is(err, context.Canceled) {
				log.Errorf("error with module %v: %v", reflect.TypeOf(module), err)
			}
			return err
		})
	}

	b.DiscordSession.AddHandler(b.handleDiscordCommand)

	guilds, err := b.DiscordSession.UserGuilds(100, "", "", false)
	log.Info("Running on servers:")
	if len(guilds) == 0 {
		log.Info("\t(none)")
	} else {
		for index := range guilds {
			guild := guilds[index]
			log.Info("\t", guild.Name, " (", guild.ID, ")")
		}
	}

	log.Infof("Join URL: https://discordapp.com/api/oauth2/authorize?scope=bot&permissions=268446720&client_id=%s", b.DiscordSession.State.User.ID)
	log.Info("Bot running. CTRL-C to exit.")

	// Await app context completion (i.e usually a SIGTERM / interrupt).
	<-ctx.Done()

	log.Info("Bot stopping...")

	// Closedown the context.
	if closeErr := g.Wait(); closeErr == nil || errors.Is(closeErr, context.Canceled) {
		log.Info("Bot stopped gracefully.")
	} else {
		log.Warn("Error during shutdown: %v", closeErr)
	}
	return

}
