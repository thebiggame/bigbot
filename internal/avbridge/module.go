package avbridge

import (
	"context"
	"github.com/andreykaipov/goobs"
	"github.com/bwmarrin/discordgo"
	"github.com/thebiggame/bigbot/internal/config"
	"github.com/thebiggame/bigbot/pkg/nodecg"
	"sync"
)

type AVBridge struct {
	discord *discordgo.Session
	// ws holds the OBS connection. ALWAYS check it is not nil before usage, and take a read on wsMtx.
	ws *goobs.Client
	// You MUST hold a read on wsMtx before using ws.
	wsMtx sync.RWMutex

	// ncg holds the NodeCG session.
	ncg *nodecg.NodeCGServer

	// The context given to us by the main bot.
	ctx *context.Context
}

func New(discord *discordgo.Session) (bridge *AVBridge, err error) {
	// Init NodeCG session
	ncg := nodecg.New(config.RuntimeConfig.AV.NodeCG.Hostname).WithKey(config.RuntimeConfig.AV.NodeCG.AuthenticationKey)
	return &AVBridge{
		discord: discord,
		ncg:     ncg,
	}, nil
}

func (mod *AVBridge) DiscordCommands() ([]*discordgo.ApplicationCommand, error) {
	return commands, nil
}

func (mod *AVBridge) Start(ctx context.Context) (err error) {
	mod.ctx = &ctx
	// goobsDaemon needs the close channel to be ready.
	goobsCtx, goobsCancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		defer wg.Done()
		// goobsDaemon is responsible for watching the goobs connection and keeping it as healthy as possible
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
