package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
)

var (
	token         string
	activeChannel = "roles"
	verbose       = false
	commandChar   = "!"
	maxUserRoles  = 5
)

func init() {
	flag.StringVar(&token, "token", "", "Bot `token` (required)")
	flag.StringVar(&activeChannel, "chan", "roles", "Channel `name` to use")
	flag.StringVar(&commandChar, "char", "!", "Command character to prefix all comamnds with")
	flag.BoolVar(&verbose, "v", false, "Verbose logging")
	flag.IntVar(&maxUserRoles, "user_maxroles", 5, "The maximum number of teams a User is allowed to join")
	flag.Parse()
	if token == "" {
		flag.Usage()
		os.Exit(1)
	}
}

func debug(v ...interface{}) {
	if verbose {
		fa := "Debug: "
		v = append([]interface{}{fa}, v...)
		log.Print(v...)
	}
}

func main() {
	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("error creating Discord session,", err)
		return
	}
	discord.AddHandler(messageCreate)
	// discord.AddHandler(ready)

	err = discord.Open()
	if err != nil {
		log.Fatal("error opening connection,", err)
		return
	}
	guilds, err := discord.UserGuilds(100, "", "", false)
	log.Print("Running on servers:")
	if len(guilds) == 0 {
		log.Print("\t(none)")
	}
	for index := range guilds {
		guild := guilds[index]
		log.Print("\t", guild.Name, " (", guild.ID, ")")
	}
	log.Print("channel name: ", activeChannel)
	log.Print("Join URL:")
	log.Print("https://discordapp.com/api/oauth2/authorize?scope=bot&permissions=268446720&client_id=", discord.State.User.ID)

	user, err := discord.User("@me")
	if err != nil {
		log.Print("Bot running. CTRL-C to exit.")
	} else {
		log.Print("Bot running as ", user.Username, "#", user.Discriminator, ". CTRL-C to exit.")
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	discord.Close()
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
			return errors.New("⚠️ You are already a member of that team!")
		}
		// Check if it's a team role, and if it is, add to the counter
		if strings.HasPrefix(role.Name, "Team:") {
			roleCount += 1
		}
	}
	if roleCount >= maxUserRoles {
		// Joining this Role would take the user over their limit
		return errors.New(fmt.Sprintf("⚠️ You are already a member of %d or more teams! Please contact an administrator if you need more.", maxUserRoles))
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
	return nil, errors.New("There was a problem creating the target role")
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		log.Print("Error getting channel:")
		log.Print(err)
		return
	}

	if strings.HasPrefix(m.Content, commandChar) {
		// it's a command character chat message
		command := m.Content[1:]
		if strings.HasPrefix(command, "jointeam") {
			if channel.Name != activeChannel {
				debug("jointeam command only works in channels with name: ", activeChannel)
				return
			}
			getRole := regexp.MustCompile(`(?:[\w]+) +(.+)`)
			regexout := getRole.FindAllStringSubmatch(m.Content, -1)
			if regexout != nil {
				roleID := regexout[0][1]
				err := validateUserCanJoinRole(s, m.Author, channel.GuildID, roleID)
				if err == nil {
					role, err := createOrReturnRole(s, channel.GuildID, roleID)
					if err == nil {
						_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintln("🙌 Joining", role.Name))
						if err == nil {
							s.GuildMemberRoleAdd(channel.GuildID, m.Author.ID, role.ID)
						}
					} else {
						s.ChannelMessageSend(m.ChannelID, err.Error())
					}
				} else {
					s.ChannelMessageSend(m.ChannelID, err.Error())
				}
			} else {
				s.ChannelMessageSend(m.ChannelID, "🤔 Please define a valid team name.")
			}
		}
	}
}
