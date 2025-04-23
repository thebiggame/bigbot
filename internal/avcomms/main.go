// Package avcomms is responsible for holding the structs of & communicating with AV related equipment at tBG Events.
package avcomms

import (
	"github.com/thebiggame/bigbot/pkg/nodecg"
	"log/slog"
	"os"
)

// logger stores the module's logger instance.
var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

func SetLogger(log *slog.Logger) {
	logger = log
}

func Init(hostname, key string) (err error) {
	if isInitialised {
		return
	}
	NodeCG = nodecg.New(hostname).WithKey(key)

	isInitialised = true
	return
}
