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
	"github.com/thebiggame/bigbot/internal/musicparty"
	"github.com/thebiggame/bigbot/internal/notifications"
	"github.com/thebiggame/bigbot/internal/shoutproxy"
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
	DiscordHandleInteraction(session *discordgo.Session, interaction *discordgo.InteractionCreate) (handled bool, err error)
	DiscordHandleMessage(session *discordgo.Session, message *discordgo.MessageCreate) (err error)
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
	modAVBridge, err := avbridge.New(b.DiscordSession)
	if err != nil {
		panic(err)
	}
	b.modules = append(b.modules, modAVBridge)

	// notifications
	modNotify, err := notifications.New(b.DiscordSession)
	if err != nil {
		panic(err)
	}
	b.modules = append(b.modules, modNotify)

	// MusicParty
	modMusic, err := musicparty.New(b.DiscordSession)
	if err != nil {
		panic(err)
	}
	b.modules = append(b.modules, modMusic)

	// ShoutProxy
	modShout, err := shoutproxy.New(b.DiscordSession)
	if err != nil {
		panic(err)
	}
	b.modules = append(b.modules, modShout)
	return b
}

func (b *BigBot) handleDiscordCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	g := new(errgroup.Group)
	for _, m := range b.modules {
		g.Go(func() error {
			handled, err := m.DiscordHandleInteraction(s, i)
			if handled {
				switch i.Type {
				case discordgo.InteractionApplicationCommand:
					log.Debugf("Module %s handled command %v", reflect.TypeOf(m).Elem().Name(), i.ApplicationCommandData().Name)
				case discordgo.InteractionModalSubmit:
					log.Debugf("Module %s handled modal response for %v", reflect.TypeOf(m).Elem().Name(), i.ModalSubmitData().CustomID)
				default:
					log.Warnf("Module %s handled command of unexpected type", reflect.TypeOf(m).Elem().Name())
				}

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

func (b *BigBot) handleDiscordMessage(s *discordgo.Session, msg *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if msg.Author.ID == s.State.User.ID {
		return
	}
	g := new(errgroup.Group)
	for _, m := range b.modules {
		g.Go(func() error {
			return m.DiscordHandleMessage(s, msg)
		})
	}
	if err := g.Wait(); err != nil {
		// Error occurred.
		log.Error(err)
		log.Debugf("Error occurred while processing: %s", msg.Message.ID)
	}
}

func (b *BigBot) registerCommands() (err error) {
	// Fetch all currently registered commands on the server.
	// this is done to avoid overwrites / deduplication.
	// We don't use ApplicationCommandOverwriteBulk because if the server's understanding of the command changes,
	// for example if role permissions change,
	// it creates a new version of that slash command causing duplication.
	guildCmds, err := b.DiscordSession.ApplicationCommands(b.DiscordSession.State.User.ID, config.RuntimeConfig.Discord.GuildID)
	if err != nil {
		return fmt.Errorf("error getting guild commands: %w", err)
	}
	// Collate all slash commands.
	for _, v := range b.modules {
		mC, err := v.DiscordCommands()
		if err != nil {
			return err
		}
		for _, cmd := range mC {
			// Write each command out individually to the guild.
			// First though, check to see whether the command already exists on the server.
			var guildCmd *discordgo.ApplicationCommand
			for _, gCmd := range guildCmds {
				if cmd.Name == gCmd.Name {
					// Match, no need to update.
					guildCmd = gCmd
					break
				}
			}
			// the cmd pointer is specifically written to here, which ensures that the ID is available to the originating slice.
			// err is scoped correctly here.
			if guildCmd != nil && guildCmd.ID != "" {
				// A command already exists on the server, just update it.
				log.Debugf("Command %s already exists on the guild, updating", cmd.Name)
				cmd, err = b.DiscordSession.ApplicationCommandEdit(b.DiscordSession.State.User.ID, config.RuntimeConfig.Discord.GuildID, guildCmd.ID, cmd)
			} else {
				// We need a new guild command.
				log.Debugf("Command %s does not exist on the guild, creating new", cmd.Name)
				cmd, err = b.DiscordSession.ApplicationCommandCreate(b.DiscordSession.State.User.ID, config.RuntimeConfig.Discord.GuildID, cmd)
			}

			if err != nil {
				return fmt.Errorf("error with discord command registration: %w", err)
			}
			// Write them to our understanding of the commands.
			// We write here with the new command knowledge (from the ApplicationCommandCreate) so that we can interact with them later (they'll have IDs).
			b.commands = append(b.commands, cmd)
		}
	}
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

	// Set appropriate intents.
	b.DiscordSession.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent)

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
	b.DiscordSession.AddHandler(b.handleDiscordMessage)

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
