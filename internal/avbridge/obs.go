package avbridge

import (
	"context"
	"github.com/andreykaipov/goobs"
	"github.com/gorilla/websocket"
	"github.com/thebiggame/bigbot/internal/config"
	"github.com/thebiggame/bigbot/internal/log"
	"time"
)

func (mod *AVBridge) goobsConnect() (err error) {
	if !mod.goobsIsConnected() {
		// GOOBS not available, connect.
		mod.wsMtx.Lock()

		var err error
		// goobs.WithLogger(config.Logger) (need a secondary logger to not pollute everything)
		mod.ws, err = goobs.New(config.RuntimeConfig.AV.OBS.Hostname, goobs.WithPassword(config.RuntimeConfig.AV.OBS.Password), goobs.WithLogger(log.Logger))
		// Not deferred as we need this back immediately.
		mod.wsMtx.Unlock()
		if err != nil {
			return err
		}
		if mod.goobsIsConnected() {
			log.LogInfo("OBS connected.")
		}
	}
	return nil
}

func (mod *AVBridge) goobsDisconnect() (err error) {
	if mod.ws != nil {
		mod.wsMtx.Lock()
		defer mod.wsMtx.Unlock()
		err := mod.ws.Disconnect()
		if err != nil {
			return err
		}
		mod.ws = nil
	}
	return nil
}

func (mod *AVBridge) goobsIsConnected() bool {
	mod.wsMtx.RLock()
	defer mod.wsMtx.RUnlock()
	if mod.ws == nil {
		return false
	}
	_, err := mod.ws.General.GetVersion()
	if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		return false
	}
	return err == nil
}

// goobsDaemon is responsible for watching the GOOBS connection on a regular basis and
// re-connecting if it seems to be unavailable for any reason.
func (mod *AVBridge) goobsDaemon(ctx context.Context) {
	err := mod.goobsConnect()
	if err != nil {
		log.LogErrf("Error connecting to OBS: %s", err)
	}
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if !mod.goobsIsConnected() {
				log.LogInfof("OBS disconnected, attempting to reconnect...")
				err = mod.goobsConnect()
				if err != nil {
					log.LogErrf("Failed to reconnect: %v", err)
				} else {
					log.LogInfof("Reconnected to OBS successfully")
				}
			}
		case <-ctx.Done():
			if mod.ws != nil {
				err := mod.goobsDisconnect()
				if err != nil {
					log.LogErrf("Error during disconnect: %v", err)
				}
				log.LogInfof("OBS disconnected gracefully.")
			}
			return
		}
	}
}
