package main

import (
	"fmt"
	"github.com/andreykaipov/goobs/api"
	"github.com/hashicorp/logutils"
	bridge_lan "github.com/thebiggame/bigbot/internal/bridge-lan"
	log2 "github.com/thebiggame/bigbot/internal/log"
	"log"
	"os"
	"strings"
)

type RunCmd struct {
	// Embed main app config (will be set during run)
	Config bridge_lan.Config `embed:"" envprefix:"BIGBRIDGE_"`
}

func (cmd *RunCmd) Run(globals *Globals) error {
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
	brInstance, err := bridge_lan.New(&cmd.Config)
	if err != nil {
		return err
	}

	return brInstance.Run()
}
