package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/thebiggame/bigbot/internal/config"
	"github.com/thebiggame/bigbot/internal/teamroles"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type BotModule interface {
	Commands() ([]*discordgo.ApplicationCommand, error)
	CommandHandlers() (map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate), error)
}

type BigBot struct {
	dSession *discordgo.Session
}

func New() *BigBot {
	bot := &BigBot{}
	return bot
}

func (b *BigBot) Run() {
	var err error

	if config.RuntimeConfig.DiscordToken == "" {
		os.Exit(1)
	}

	b.dSession, err = discordgo.New("Bot " + config.RuntimeConfig.DiscordToken)
	if err != nil {
		log.Fatal("error creating Discord session,", err)
	}

	b.dSession.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	err = b.dSession.Open()
	if err != nil {
		log.Fatal("error opening connection,", err)
	}
	defer b.dSession.Close()

	log.Println("adding commands...")
	var commands []*discordgo.ApplicationCommand
	commandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){}

	// teamRoles
	commands = append(commands, teamroles.Commands...)
	for k, v := range teamroles.CommandHandlers {
		commandHandlers[k] = v
	}

	b.dSession.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := b.dSession.ApplicationCommandCreate(b.dSession.State.User.ID, config.RuntimeConfig.DiscordGuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	guilds, err := b.dSession.UserGuilds(100, "", "", false)
	log.Print("Running on servers:")
	if len(guilds) == 0 {
		log.Print("\t(none)")
	}
	for index := range guilds {
		guild := guilds[index]
		log.Print("\t", guild.Name, " (", guild.ID, ")")
	}
	log.Print("Join URL:")
	log.Print("https://discordapp.com/api/oauth2/authorize?scope=bot&permissions=268446720&client_id=", b.dSession.State.User.ID)

	log.Print("Bot running. CTRL-C to exit.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

}
