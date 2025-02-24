package helpers

import (
	"github.com/thebiggame/bigbot/internal/config"
	"log"
)

func LogDebug(v ...interface{}) {
	if config.RuntimeConfig.Verbose {
		fa := "Debug: "
		v = append([]interface{}{fa}, v...)
		log.Print(v...)
	}
}
