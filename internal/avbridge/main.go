package avbridge

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"github.com/thebiggame/bigbot/internal/avcomms"
	"sync"
)

type AVBridge struct {
	discord *discordgo.Session

	// The context given to us by the main bot.
	ctx *context.Context
}

func New(discord *discordgo.Session) (bridge *AVBridge, err error) {
	// InitOld AV session handlers in avcomms
	// Unbind this tight integration perhaps?
	err = avcomms.InitOld()
	if err != nil {
		return nil, err
	}
	return &AVBridge{
		discord: discord,
	}, nil
}

func (mod *AVBridge) DiscordCommands() ([]*discordgo.ApplicationCommand, error) {
	return commands, nil
}

func (mod *AVBridge) Start(ctx context.Context) (err error) {
	mod.ctx = &ctx
	// TODO This feels like the wrong place to initialise avcomms.
	// goobsDaemon needs the close channel to be ready.
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
