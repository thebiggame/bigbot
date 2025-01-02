package main

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/jessevdk/go-flags"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
)

var (
	runtimeOptions struct {
		DiscordToken   string `short:"t" long:"token" description:"Discord bot token" required:"true" env:"DISCORD_TOKEN"`
		DiscordGuildID string `long:"guildID" description:"Discord guild ID to monitor" default:"" env:"DISCORD_GUILD"`
		Verbose        bool   `short:"v" long:"verbose" description:"Show verbose debug information"`
		MaxUserRoles   int    `long:"maxUserRoles" default:"5" description:"Maximum number of teams a User can join"`
		RemoveCommands bool   `long:"removeCommands" description:"Remove commands on shutdown"`
	}

	dSession *discordgo.Session
)

func init() {
	_, err := flags.Parse(&runtimeOptions)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	RunBot()
}

func RunBot() {
	var err error

	if runtimeOptions.DiscordToken == "" {
		os.Exit(1)
	}

	dSession, err = discordgo.New("Bot " + runtimeOptions.DiscordToken)
	if err != nil {
		log.Fatal("error creating Discord session,", err)
	}

	dSession.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	dSession.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	err = dSession.Open()
	if err != nil {
		log.Fatal("error opening connection,", err)
	}
	defer dSession.Close()

	log.Println("adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := dSession.ApplicationCommandCreate(dSession.State.User.ID, runtimeOptions.DiscordGuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	guilds, err := dSession.UserGuilds(100, "", "", false)
	log.Print("Running on servers:")
	if len(guilds) == 0 {
		log.Print("\t(none)")
	}
	for index := range guilds {
		guild := guilds[index]
		log.Print("\t", guild.Name, " (", guild.ID, ")")
	}
	log.Print("Join URL:")
	log.Print("https://discordapp.com/api/oauth2/authorize?scope=bot&permissions=268446720&client_id=", dSession.State.User.ID)

	log.Print("Bot running. CTRL-C to exit.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

}

func debug(v ...interface{}) {
	if runtimeOptions.Verbose {
		fa := "Debug: "
		v = append([]interface{}{fa}, v...)
		log.Print(v...)
	}
}

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "team",
		Description: "Manage your membership of LAN teams",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "join",
				Description: "Join a team",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "team-name",
						Description: "Name of the team you wish to join.",
						Required:    true,
					},
				},
			},
		},
	},
}

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"team": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		options := i.ApplicationCommandData().Options
		content := ""

		// As you can see, names of subcommands (nested, top-level)
		// and subcommand groups are provided through the arguments.
		switch options[0].Name {

		case "join":
			// OOB check
			if len(options[0].Options) < 0 {
				content = "🤔 Please provide a team name."
				break
			}
			if i.Interaction.GuildID == "" {
				content = "😡 This command can only be used in a server."
				break
			}
			roleID := options[0].Options[0].StringValue()
			err := validateUserCanJoinRole(s, i.Interaction.Member.User, i.GuildID, roleID)
			if err != nil {
				content = err.Error()
				break
			}
			role, err := createOrReturnRole(s, i.GuildID, roleID)
			if err != nil {
				content = err.Error()
				break
			}
			err = s.GuildMemberRoleAdd(i.GuildID, i.Interaction.Member.User.ID, role.ID)
			if err != nil {
				content = err.Error()
				break
			}
			content = fmt.Sprintln("🙌 Joined", role.Name)
		default:
			content = "😶 Please use a sub-command."
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	},
}

func validateUserCanJoinRole(s *discordgo.Session, u *discordgo.User, guild string, targetRole string) (err error) {
	// This function validates that the given GuildMember satisfies the following rules:
	// - is not already assigned to more than 5 Team Roles
	// - is not already assigned to the given targetRole
	var roleCount int
	member, err := s.GuildMember(guild, u.ID)
	if err != nil {
		return err
	}
	for _, v := range member.Roles {
		role, err := s.State.Role(guild, v)
		if err != nil {
			return err
		}
		getRoleName := regexp.MustCompile(`(?i)(?:team):* ?(.*)`)
		roleName := getRoleName.FindAllStringSubmatch(role.Name, -1)
		// role names get normalized to lower case during the lookup only
		if roleName != nil && strings.ToLower(roleName[0][1]) == strings.ToLower(targetRole) {
			// The Member is already part of the given GuildRole!
			return errors.New("⚠️ You are already a member of that team")
		}
		// Check if it's a team role, and if it is, add to the counter
		if strings.HasPrefix(role.Name, "Team:") {
			roleCount += 1
		}
	}
	if roleCount >= runtimeOptions.MaxUserRoles {
		// Joining this Role would take the user over their limit
		return errors.New(fmt.Sprintf("⚠️ You are already a member of %d or more teams! Please contact an administrator if you need more", runtimeOptions.MaxUserRoles))
	}

	// Succ(ess)
	return nil

}

func createOrReturnRole(s *discordgo.Session, guild string, rname string) (v *discordgo.Role, err error) {
	roles, err := s.GuildRoles(guild)
	getRole := regexp.MustCompile(`(?i)(?:team):*`)
	if !getRole.MatchString(rname) {
		rname = fmt.Sprintln("Team:", rname)
	}
	rname = strings.Replace(rname, "\n", "", -1)
	if err == nil {
		for _, v := range roles {
			// role names get normalized to lower case during the lookup only
			if strings.ToLower(v.Name) == strings.ToLower(rname) {
				log.Print("tying", rname, "to old role", v.Name)
				return v, nil
			}
		}
		// couldn't find the role in our list, create it
		log.Print("creating new role ", rname)
		var rColour = 8290694
		var rHoist = true
		var rMentionable = true
		var rPerms int64 = 0
		rParams := discordgo.RoleParams{
			Name:         rname,
			Color:        &rColour,
			Hoist:        &rHoist,
			Permissions:  &rPerms,
			Mentionable:  &rMentionable,
			UnicodeEmoji: nil,
			Icon:         nil,
		}
		role, err := s.GuildRoleCreate(guild, &rParams)
		return role, err
	}
	return nil, errors.New("problem creating the target role")
}
