package main

import (
	"github.com/thebiggame/bigbot/internal/bot"
	"github.com/thebiggame/bigbot/internal/config"
)

type RunCmd struct {
	// Embed main app config (will be set during run)
	Config config.Config `embed:"" envprefix:"BIGBOT_"`
}

func (cmd *RunCmd) Run(globals *Globals) error {
	// Bind config to global app config struct
	config.RuntimeConfig = cmd.Config

	botInstance, err := bot.New()
	if err != nil {
		return err
	}

	return botInstance.Run()
}
