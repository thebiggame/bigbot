package config

// Config defines the format of the application configuration.
type Config struct {
	Bridge struct {
		Enabled bool         `long:"enabled" help:"Enable the BIGbot -> Bridge Server" default:"false" env:"ENABLED"`
		Address string       `long:"listen" help:"Listen address and port" default:"localhost:8080" env:"LISTEN"`
		Key     SecretString `long:"key" help:"BIGbot authentication key" env:"KEY"`
	} `prefix:"bridge." embed:"" envprefix:"BRIDGE_"`
	Discord struct {
		Token         SecretString `short:"t" long:"token" help:"Discord bot token" required:"" env:"TOKEN"`
		GuildID       string       `long:"guildID" help:"Discord guild ID to monitor" default:"" env:"GUILD"`
		Announcements struct {
			ChannelID string `json:"channelID" help:"Channel ID" default:"" env:"CHANNEL"`
		} `prefix:"announcements." embed:"" envprefix:"ANNOUNCEMENTS_"`
		Permissions struct {
			CrewRole string `help:"If a user is a member of this role ID, treat them as Crew." default:"" env:"ROLE_CREW"`
		} `prefix:"permissions." embed:"" envprefix:"PERMISSIONS_"`
		Shoutbox struct {
			ChannelID string `json:"channelID" help:"Channel ID" default:"" env:"CHANNEL"`
		} `prefix:"shoutbox." embed:"" envprefix:"SHOUTBOX_"`
	} `prefix:"discord." embed:"" envprefix:"DISCORD_"`
	AV struct {
		NodeCG struct {
			BundleName string `long:"bundle" help:"NodeCG bundle name" default:"thebiggame" env:"BUNDLE"`
		} `prefix:"nodecg." embed:"" envprefix:"NODECG_"`
	} `prefix:"av." embed:"" envprefix:"AV_"`
	Teams struct {
		MaxUserTeams int `long:"maxUserRoles" default:"5" help:"Maximum number of teams a User can join" env:"MAX_USER_ROLES"`
	} `prefix:"teams." embed:""`
	RemoveCommands bool `long:"removeCommands" help:"Remove commands on shutdown" env:"COMMANDS_REMOVE"`
}

// RuntimeConfig holds the current state of the configuration for BIGbot.
// Stored here instead of in the main bot struct for ease of access.
var RuntimeConfig Config
