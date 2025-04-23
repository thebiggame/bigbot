package main

import (
	bridge_lan "github.com/thebiggame/bigbot/internal/bridge-lan"
	"github.com/thebiggame/bigbot/internal/config"
)

type BridgeCmd struct {
	// Embed main app config (will be set during run)
	Config config.BridgeConfig `embed:"" envprefix:"BIGBRIDGE_"`
}

func (cmd *BridgeCmd) Run(globals *Globals) error {
	brInstance, err := bridge_lan.New(&cmd.Config)
	if err != nil {
		return err
	}

	return brInstance.Run()
}
