package main

import (
	"github.com/thebiggame/bigbot/internal/bot"
	"github.com/thebiggame/bigbot/internal/config"
	log2 "github.com/thebiggame/bigbot/internal/log"
	"log/slog"
	"os"
	"strings"
)

type RunCmd struct {
	// Embed main app config (will be set during run)
	Config config.Config `embed:"" envprefix:"BIGBOT_"`
}

func (cmd *RunCmd) Run(globals *Globals) error {
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

	return botInstance.Run()
}
