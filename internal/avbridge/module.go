package avbridge

import (
	"context"
	"github.com/andreykaipov/goobs"
	"github.com/bwmarrin/discordgo"
	"sync"
)

type AVBridge struct {
	discord *discordgo.Session
	// ws holds the OBS connection. ALWAYS check it is not nil before usage, and take a read on wsMtx.
	ws *goobs.Client
	// You MUST hold a read on wsMtx before using ws.
	wsMtx sync.RWMutex
}

func New(discord *discordgo.Session) (bridge *AVBridge, err error) {
	return &AVBridge{
		discord: discord,
	}, nil
}

func (mod *AVBridge) DiscordCommands() ([]*discordgo.ApplicationCommand, error) {
	return commands, nil
}

func (mod *AVBridge) Start(ctx context.Context) (err error) {
	// goobsDaemon needs the close channel to be ready.
	goobsCtx, goobsCancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		defer wg.Done()
		mod.goobsDaemon(goobsCtx)
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
