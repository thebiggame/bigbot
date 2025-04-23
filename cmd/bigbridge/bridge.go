package main

import (
	bridge_lan "github.com/thebiggame/bigbot/internal/bridge-lan"
	"github.com/thebiggame/bigbot/internal/config"
	log2 "github.com/thebiggame/bigbot/internal/log"
	"strings"
)

type RunCmd struct {
	// Embed main app config (will be set during run)
	Config config.BridgeConfig `embed:"" envprefix:"BIGBRIDGE_"`
}

func (cmd *RunCmd) Run(globals *Globals) error {
	// Configure logging
	logLevelNormalised := strings.ToUpper(globals.LogLevel)
	if logLevelNormalised == "TRACE" {
		log2.Level.Set(log2.LevelTrace)
	}

	brInstance, err := bridge_lan.New(&cmd.Config)
	if err != nil {
		return err
	}

	return brInstance.Run()
}
