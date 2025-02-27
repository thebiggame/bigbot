// Package config provides the configuration state and utilities for working with the global bot configuration.
package config

import "github.com/spf13/viper"

// Config defines the format of the application configuration.
type Config struct {
	Discord struct {
		Token   string `short:"t" long:"token" description:"Discord bot token" required:"true" env:"BIGBOT_DISCORD_TOKEN"`
		GuildID string `long:"guildID" description:"Discord guild ID to monitor" default:"" env:"BIGBOT_DISCORD_GUILD"`
	}
	LogLevel       string `short:"log" long:"loglevel" description:"Set log level"`
	MaxUserRoles   int    `long:"maxUserRoles" default:"5" description:"Maximum number of teams a User can join"`
	RemoveCommands bool   `long:"removeCommands" description:"Remove commands on shutdown"`
}

// RuntimeConfig holds the current state of the configuration for BIGbot.
// Stored here instead of in the main bot struct for ease of access.
var RuntimeConfig Config

// BindViperConfig binds the appropriate Environment Variables to the configuration.
func BindViperConfig() {
	// Very yucky, would prefer to not have to do this but apparently this isn't possible to work around with Viper.
	// Shame, because Cobra is cool otherwise.
	viper.BindEnv("Discord.Token", "BIGBOT_DISCORD_TOKEN")
	viper.BindEnv("Discord.GuildID", "BIGBOT_GUILD_ID")
	viper.BindEnv("MaxUserRoles", "BIGBOT_MAX_USER_ROLES")
	viper.BindEnv("RemoveCommands", "BIGBOT_REMOVE_COMMANDS")
	return
}
