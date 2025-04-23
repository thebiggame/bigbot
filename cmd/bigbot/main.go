package main

import (
	"github.com/alecthomas/kong"
	"github.com/alecthomas/kong-yaml"
	log2 "github.com/thebiggame/bigbot/internal/log"
	"log/slog"
	"os"
	"strings"
)

type Globals struct {
	Config   kong.ConfigFlag `help:"Location of config" env:"CONFIG" type:"path"`
	LogLevel string          `short:"l" help:"Set the logging level (TRACE|DEBUG|INFO|WARN|ERROR|FATAL)" enum:"TRACE,DEBUG,INFO,WARN,ERROR,FATAL" env:"LOG_LEVEL" default:"INFO"`
}

type CLI struct {
	Globals `envprefix:"BIGBOT_"`

	Run    RunCmd    `cmd:"run" help:"Run BIGbot (."`
	Bridge BridgeCmd `cmd:"bridge" help:"Run BIGbridge (the event client)."`
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli,
		kong.Name("bigbot"),
		kong.Description("The tBG Discord bot."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Configuration(kongyaml.Loader, "/etc/bigbot.yaml", "./bigbot.yaml"))

	// Configure logging
	var logLevel slog.Level
	logLevelNormalised := strings.ToUpper(cli.LogLevel)
	if level, ok := log2.Levels[logLevelNormalised]; ok {
		logLevel = level
	} else {
		panic("unknown log level " + logLevelNormalised)
	}

	log2.Level.Set(logLevel)

	log2.Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:       log2.Level,
		ReplaceAttr: log2.Replacer,
	},
	))

	err := ctx.Run(&cli.Globals)
	ctx.FatalIfErrorf(err)
}
