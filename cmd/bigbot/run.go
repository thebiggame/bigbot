package main

import (
	"errors"
	"fmt"
	"github.com/andreykaipov/goobs/api"
	"github.com/hashicorp/logutils"
	"github.com/thebiggame/bigbot/internal/bot"
	"github.com/thebiggame/bigbot/internal/config"
	log2 "github.com/thebiggame/bigbot/internal/log"
	"log"
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
	logFlags := log.Ltime
	logLevelNormalised := strings.ToUpper(globals.LogLevel)
	if logLevelNormalised == "TRACE" {
		logFlags |= log.Llongfile
	}
	log2.Logger = log.New(
		&logutils.LevelFilter{
			Levels:   []logutils.LogLevel{"TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"},
			MinLevel: logutils.LogLevel(logLevelNormalised),
			Writer: api.LoggerWithWrite(func(p []byte) (int, error) {
				return os.Stderr.WriteString(fmt.Sprintf("\033[36m%s\033[0m", p))
			}),
		},
		"",
		logFlags,
	)
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
