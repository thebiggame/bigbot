package avbridge

import (
	"context"
	"fmt"
	"github.com/andreykaipov/goobs"
	"github.com/bwmarrin/discordgo"
	"github.com/thebiggame/bigbot/internal/config"
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

func (mod *AVBridge) Start(ctx context.Context) (err error) {
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	cmds, err := mod.discord.ApplicationCommandBulkOverwrite(mod.discord.State.User.ID, config.RuntimeConfig.Discord.GuildID, commands)
	if err != nil {
		// This shouldn't happen. Bail out
		return fmt.Errorf("error creating commands: %w", err)
	}
	copy(registeredCommands, cmds)

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
