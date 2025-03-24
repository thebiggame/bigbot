package notifications

import (
	"context"
	"github.com/bwmarrin/discordgo"
)

type Notifications struct {
	discord *discordgo.Session

	// The context given to us by the main bot.
	ctx *context.Context
}

func New(discord *discordgo.Session) (mod *Notifications, err error) {
	return &Notifications{
		discord: discord,
	}, nil
}

var commands = []*discordgo.ApplicationCommand{
	{
		Name:                     "notify",
		Description:              "🛎️ Send & manage notifications (you must be a crew member)",
		DefaultMemberPermissions: &defaultCrewCommandPermissions,
		DMPermission:             &defaultCrewCommandDMPermissions,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "alert",
				Description: "🔔 Sound an Alert on the AV system.",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionBoolean,
						Name:        "flair",
						Description: "Whether the alert should arrive with 'flair'. WARNING - this makes noise!",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "name",
						Description: "A short description of why you want people's attention.",
						Required:    true,
						MaxLength:   40,
					},
				},
			},
			{
				Name:        "alert-end",
				Description: "🔕 End the Alert early.",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
			{
				Name:        "announcement",
				Description: "🔔 Open the Announcement modal to create a new Announcement. (Submitting this makes noise!)",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
			{
				Name:        "announcement-end",
				Description: "🔕 End the Announcement (return to normal service).",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
		},
	},
}

func (mod *Notifications) DiscordCommands() ([]*discordgo.ApplicationCommand, error) {
	return commands, nil
}

func (mod *Notifications) Start(ctx context.Context) (err error) {
	mod.ctx = &ctx
	// This module simply registers handlers, and does not need to run continuously (so we don't need the context)
	return ctx.Err()
}
