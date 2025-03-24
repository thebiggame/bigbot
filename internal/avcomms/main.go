// Package avcomms is responsible for holding the structs of & communicating with AV related equipment at tBG Events.
// It is relatively tightly bound to the avbridge package - but factored out to allow other modules to directly communicate with AV if necessary.
package avcomms

import (
	"github.com/thebiggame/bigbot/internal/config"
	"github.com/thebiggame/bigbot/pkg/nodecg"
)

func Init() (err error) {
	if isInitialised {
		return
	}
	NodeCG = nodecg.New(config.RuntimeConfig.AV.NodeCG.Hostname).WithKey(config.RuntimeConfig.AV.NodeCG.AuthenticationKey)

	isInitialised = true
	return
}
