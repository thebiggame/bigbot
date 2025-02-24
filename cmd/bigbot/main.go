package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/thebiggame/bigbot/internal/bot"
	"github.com/thebiggame/bigbot/internal/config"
	"log"
)

func init() {
	_, err := flags.Parse(&config.RuntimeConfig)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	bot.New().Run()
}
