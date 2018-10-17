package main

import (
  "flag"
  "log"
  "strings"
  "regexp"
  "os"
  "os/signal"
  "syscall"

  "github.com/bwmarrin/discordgo"
  "errors"
  "fmt"
)

var (
  token string
  activeChannel = "roles"
  verbose = false
  command_char = "!"
)

func init() {
  flag.StringVar(&token, "token", "", "Bot `token` (required)")
  flag.StringVar(&activeChannel, "chan", "roles", "Channel `name` to use")
  flag.StringVar(&command_char, "char", "!", "Command character to prefix all comamnds with")
  flag.BoolVar(&verbose, "v", false, "Verbose logging")
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

func main()  {
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
  guilds, err := discord.UserGuilds(100, "", "")
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

func createOrReturnRole(s *discordgo.Session, guild string, rname string) (v *discordgo.Role, err error) {
  roles, err := s.GuildRoles(guild)
  // I am way too lazy for regex here
  if !strings.HasPrefix(rname, "Team:") || !strings.HasPrefix(rname, "team:") || !strings.HasPrefix(rname, "Team") || !strings.HasPrefix(rname, "team") {
    rname = fmt.Sprintln("Team: ", rname)
  }
  rname = strings.Replace(rname, "\n", "", -1)
  if err == nil {
    for _, v := range roles {
      if v.Name == rname {
        return v, nil
      }
    }
    // couldn't find the role in our list, create it
    role, err := s.GuildRoleCreate(guild)
    if err == nil {
      // Patch the role
      return s.GuildRoleEdit(guild, role.ID, rname, 8290694, true, 0, true)
    }
  }
  return nil, errors.New("there was a problem creating the target role")
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

  if strings.HasPrefix(m.Content, "!") {
    // it's a command character chat message
    command := m.Content[1:]
    if strings.HasPrefix(command, "jointeam") {
      if channel.Name != activeChannel {
        debug("jointeam command only works in channels with name: ", activeChannel)
        return
      }
      getRole := regexp.MustCompile(`(?:[\w]+) ?(.+)`)
      regexout := getRole.FindAllStringSubmatch(m.Content, -1)
      if regexout != nil {
        roleID := regexout[0][1]
        text := []string{}
        if regexout[0][1] != "" {
          log.Print("registering ", roleID)
          text = []string{"Joining team: `", roleID, "`\n"}
        } else {
          log.Print("registering ", roleID, ": ")
          text = []string{"Joining team: `", roleID, "`\n"}
        }

        _, err := s.ChannelMessageSend(m.ChannelID, strings.Join(text, ""))
        if err == nil {
          role, err := createOrReturnRole(s, channel.GuildID, roleID)
          if err == nil {
            s.GuildMemberRoleAdd(channel.GuildID, m.Author.ID, role.ID)
          }
        }
        }
      }
  }
}
