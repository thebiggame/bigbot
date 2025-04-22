package avbridge

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"log/slog"
)

type AVBridge struct {
	discord *discordgo.Session

	// The logger for this module.
	logger *slog.Logger

	// The context given to us by the main bot.
	ctx *context.Context
}

func New(discord *discordgo.Session) (bridge *AVBridge, err error) {
	return &AVBridge{
		discord: discord,
	}, nil
}

func (mod *AVBridge) SetLogger(logger *slog.Logger) {
	mod.logger = logger
}

func (mod *AVBridge) DiscordCommands() ([]*discordgo.ApplicationCommand, error) {
	return commands, nil
}

func (mod *AVBridge) Start(ctx context.Context) (err error) {
	mod.ctx = &ctx
	return ctx.Err()
}
