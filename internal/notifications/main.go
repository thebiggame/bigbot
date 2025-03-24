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

func (mod *Notifications) DiscordCommands() ([]*discordgo.ApplicationCommand, error) {
	return commands, nil
}

func (mod *Notifications) Start(ctx context.Context) (err error) {
	mod.ctx = &ctx
	// This module simply registers handlers, and does not need to run continuously (so we don't need the context)
	return ctx.Err()
}
