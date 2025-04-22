package avbridge

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"github.com/thebiggame/bigbot/internal/avcomms"
	"log/slog"
	"sync"
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
	// FIXME Deprecated; to be moved to bridge-lan.
	goobsCtx, goobsCancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		defer wg.Done()
		// goobsDaemon is responsible for watching the goobs connection and keeping it as healthy as possible
		avcomms.OBSDaemon(goobsCtx)
	}()

	for {
		select {
		// Spinloop here to make sure that we stay alive long enough for the context to get torn down properly.
		case <-ctx.Done():
			goobsCancel()
			wg.Wait()
			return ctx.Err()
		}
	}
}
