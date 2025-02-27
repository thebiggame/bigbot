package bot

import (
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/thebiggame/bigbot/internal/avbridge"
	"github.com/thebiggame/bigbot/internal/config"
	log "github.com/thebiggame/bigbot/internal/log"
	"github.com/thebiggame/bigbot/internal/teamroles"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
)

type BotModule interface {
	Start(context context.Context) error
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
	var wg sync.WaitGroup
	errorChan := make(chan error, len(b.modules))
	for _, m := range b.modules {
		wg.Add(1)
		go func(mod *BotModule) {
			defer wg.Done()
			handled, err := m.HandleDiscordCommand(s, i)
			if err != nil {
				errorChan <- err
				return
			}
			if handled {
				log.LogDebugf("Module %s handled command %v", reflect.TypeOf(mod), i.ApplicationCommandData().Name)
			}
		}(&m)
	}
	wg.Wait()
	for range errorChan {
		err := <-errorChan
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: err.Error(),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

}

func (b *BigBot) Run() (err error) {
	if config.RuntimeConfig.Discord.Token == "" {
		return errors.New("no discord token provided")
	}

	b.DiscordSession.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.LogInfof("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	err = b.DiscordSession.Open()
	if err != nil {
		return fmt.Errorf("error opening discord connection: %w", err)
	}
	defer b.DiscordSession.Close()

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
				log.LogErrf("error with module %v: %v", reflect.TypeOf(module), err)
			}
			return err
		})
	}

	b.DiscordSession.AddHandler(b.handleDiscordCommand)

	guilds, err := b.DiscordSession.UserGuilds(100, "", "", false)
	log.LogInfo("Running on servers:")
	if len(guilds) == 0 {
		log.LogInfo("\t(none)")
	} else {
		for index := range guilds {
			guild := guilds[index]
			log.LogInfo("\t", guild.Name, " (", guild.ID, ")")
		}
	}

	log.LogInfof("Join URL: https://discordapp.com/api/oauth2/authorize?scope=bot&permissions=268446720&client_id=%s", b.DiscordSession.State.User.ID)
	log.LogInfo("Bot running. CTRL-C to exit.")

	// Await app context completion (i.e usually a SIGTERM / interrupt).
	<-ctx.Done()

	log.LogInfo("Bot stopping...")

	// Closedown the context.
	if closeErr := g.Wait(); closeErr == nil || errors.Is(closeErr, context.Canceled) {
		log.LogInfo("Bot stopped gracefully.")
	} else {
		log.LogWarn("Error during shutdown: %v", closeErr)
	}
	return

}
