package main

import (
	"errors"
	"github.com/thebiggame/bigbot/internal/bot"
	"github.com/thebiggame/bigbot/internal/config"
	log2 "github.com/thebiggame/bigbot/internal/log"
	"log/slog"
	"os"
	"strings"
)

type RunCmd struct {
	ServeLAN bool `group:"serve" help:"Serve the LAN portion of the bot."`
	ServeWAN bool `group:"serve" help:"Serve the WAN portion of the bot."`

	// Embed main app config (will be set during run)
	Config config.Config `embed:"" envprefix:"BIGBOT_"`
}

func (cmd *RunCmd) Run(globals *Globals) error {
	if !cmd.ServeLAN && !cmd.ServeWAN {
		// No run type is set, error.
		return errors.New("you must use at least one of --serve-wan, --serve-lan")
	}
	// Bind config to global app config struct
	config.RuntimeConfig = cmd.Config

	// Configure logging
	var logLevel slog.Level
	logLevelNormalised := strings.ToUpper(globals.LogLevel)
	if logLevelNormalised == "TRACE" {
		log2.Level.Set(log2.LevelTrace)
	}

	log2.Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	},
	))

	botInstance, err := bot.New()
	if err != nil {
		return err
	}

	if cmd.ServeLAN {
		botInstance = botInstance.WithLANModules()
	}
	if cmd.ServeWAN {
		botInstance = botInstance.WithWANModules()
	}
	return botInstance.Run()
}
