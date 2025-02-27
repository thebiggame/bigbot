package run

import (
	"fmt"
	"github.com/andreykaipov/goobs/api"
	"github.com/hashicorp/logutils"
	"github.com/spf13/cobra"
	"github.com/thebiggame/bigbot/internal/bot"
	log2 "github.com/thebiggame/bigbot/internal/log"
	"log"
	"os"
	"strings"
)

var (
	serveLAN bool
	serveWAN bool
)

var Cmd = &cobra.Command{
	Use:   "run",
	Short: "Run the bot.",
	Run:   Run,
}

func init() {
	Cmd.Flags().BoolVar(&serveLAN, "lan", false, "Serve the LAN portion of the bot.")
	Cmd.Flags().BoolVar(&serveWAN, "wan", false, "Serve the WAN portion of the bot.")
	Cmd.MarkFlagsOneRequired("lan", "wan")
}

func Run(cmd *cobra.Command, args []string) {
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
	botInstance := bot.New()
	if serveLAN {
		botInstance = botInstance.WithLANModules()
	}
	if serveWAN {
		botInstance = botInstance.WithWANModules()
	}

	cobra.CheckErr(botInstance.Run())
}
