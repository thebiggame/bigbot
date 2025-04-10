package musicparty

import (
	"context"
	"github.com/bwmarrin/discordgo"
)

type MusicParty struct {
	discord *discordgo.Session

	// The context given to us by the main bot.
	ctx *context.Context
}

func New(discord *discordgo.Session) (mod *MusicParty, err error) {
	return &MusicParty{
		discord: discord,
	}, nil
}

func (mod *MusicParty) DiscordCommands() ([]*discordgo.ApplicationCommand, error) {
	return commands, nil
}

func (mod *MusicParty) Start(ctx context.Context) (err error) {
	mod.ctx = &ctx
	// This module simply registers handlers, and does not need to run continuously (so we don't need the context)
	return ctx.Err()
}
