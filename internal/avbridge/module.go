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
	for i, v := range commands {
		cmd, err := mod.discord.ApplicationCommandCreate(mod.discord.State.User.ID, config.RuntimeConfig.Discord.GuildID, v)
		if err != nil {
			// This shouldn't happen. Bail out
			return fmt.Errorf("error creating command '%v': %w", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	// goobsDaemon needs the close channel to be ready.
	goobsCtx, cancel := context.WithCancel(ctx)

	go mod.goobsDaemon(goobsCtx)

	for {
		select {
		// Spinloop here to make sure that we stay alive long enough for the context to get torn down properly.
		case <-ctx.Done():
			cancel()
			<-goobsCtx.Done()
			return ctx.Err()
		}
	}
}
