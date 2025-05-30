package shoutproxy

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"os"
)

type ShoutProxy struct {
	discord *discordgo.Session

	// The context given to us by the main bot.
	ctx *context.Context
}

// logger stores the module's logger instance.
var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

func New(discord *discordgo.Session) (mod *ShoutProxy, err error) {
	return &ShoutProxy{
		discord: discord,
	}, nil
}

func (mod *ShoutProxy) SetLogger(log *slog.Logger) {
	logger = log
}

func (mod *ShoutProxy) DiscordCommands() ([]*discordgo.ApplicationCommand, error) {
	return []*discordgo.ApplicationCommand{}, nil
}

func (mod *ShoutProxy) Start(ctx context.Context) (err error) {
	mod.ctx = &ctx
	// This module simply registers handlers, and does not need to run continuously (so we don't need the context)
	return ctx.Err()
}
