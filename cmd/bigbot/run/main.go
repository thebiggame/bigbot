package run

import (
	"fmt"
	"github.com/andreykaipov/goobs/api"
	"github.com/hashicorp/logutils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	logFlags := log.Ltime
	logLevel := strings.ToUpper(viper.GetString("log.level"))
	if logLevel == "TRACE" {
		logFlags |= log.Llongfile
	}
	log2.Logger = log.New(
		&logutils.LevelFilter{
			Levels:   []logutils.LogLevel{"TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"},
			MinLevel: logutils.LogLevel(strings.ToUpper(viper.GetString("log.level"))),
			Writer: api.LoggerWithWrite(func(p []byte) (int, error) {
				return os.Stderr.WriteString(fmt.Sprintf("\033[36m%s\033[0m", p))
			}),
		},
		"",
		logFlags,
	)
	botInstance, err := bot.New()
	cobra.CheckErr(err)

	if serveLAN {
		botInstance = botInstance.WithLANModules()
	}
	if serveWAN {
		botInstance = botInstance.WithWANModules()
	}

	cobra.CheckErr(botInstance.Run())
}
