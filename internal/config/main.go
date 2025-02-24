package config

type Config struct {
	DiscordToken   string `short:"t" long:"token" description:"Discord bot token" required:"true" env:"DISCORD_TOKEN"`
	DiscordGuildID string `long:"guildID" description:"Discord guild ID to monitor" default:"" env:"DISCORD_GUILD"`
	Verbose        bool   `short:"v" long:"verbose" description:"Show verbose debug information"`
	MaxUserRoles   int    `long:"maxUserRoles" default:"5" description:"Maximum number of teams a User can join"`
	RemoveCommands bool   `long:"removeCommands" description:"Remove commands on shutdown"`
}

var RuntimeConfig Config
