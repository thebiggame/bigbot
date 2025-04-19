package main

import (
	"github.com/alecthomas/kong"
	"github.com/alecthomas/kong-yaml"
)

type Globals struct {
	Config   kong.ConfigFlag `help:"Location of config" env:"CONFIG" type:"path"`
	LogLevel string          `short:"l" help:"Set the logging level (debug|info|warn|error|fatal)" env:"LOG_LEVEL" default:"info"`
}

type CLI struct {
	Globals `envprefix:"BIGBRIDGE_"`

	Run RunCmd `cmd:"" help:"Run the event-to-BIGbot bridge server."`
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli,
		kong.Name("bigbridge"),
		kong.Description("The event-to-BIGbot bridge. Designed to run inside the LAN party."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Configuration(kongyaml.Loader, "/etc/bigbridge.yaml", "./bigbridge.yaml"))
	err := ctx.Run(&cli.Globals)
	ctx.FatalIfErrorf(err)
}
