package main

import (
	"fmt"
	"github.com/andreykaipov/goobs/api"
	"github.com/hashicorp/logutils"
	"github.com/jessevdk/go-flags"
	"github.com/thebiggame/bigbot/internal/bot"
	"github.com/thebiggame/bigbot/internal/config"
	log2 "github.com/thebiggame/bigbot/internal/log"
	"log"
	"os"
	"strings"
)

func init() {
	_, err := flags.Parse(&config.RuntimeConfig)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	log2.SetLogger(log.New(
		&logutils.LevelFilter{
			Levels:   []logutils.LogLevel{"TRACE", "DEBUG", "INFO", "ERROR", "FATAL"},
			MinLevel: logutils.LogLevel(strings.ToUpper(os.Getenv("BIGBOT_LOG"))),
			Writer: api.LoggerWithWrite(func(p []byte) (int, error) {
				return os.Stderr.WriteString(fmt.Sprintf("\033[36m%s\033[0m", p))
			}),
		},
		"",
		log.Ltime|log.Lshortfile,
	))
	bot.New().WithWANModules().WithLANModules().Run()
}
