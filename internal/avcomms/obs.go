package avcomms

import (
	"context"
	"github.com/andreykaipov/goobs"
	"github.com/gorilla/websocket"
	"github.com/thebiggame/bigbot/internal/config"
	"github.com/thebiggame/bigbot/internal/log"
	"sync"
	"time"
)

var (
	// isInitialised defines whether this module has been set up (avoids double-setup race issues).
	isInitialised bool

	// OBS holds the OBS connection. ALWAYS check it is not nil before usage, and take a read on wsMtx.
	OBS *goobs.Client

	// OBSMtx MUST be held before using OBS.
	OBSMtx sync.RWMutex
)

func goobsConnect() (err error) {
	if GoobsIsConnected() {
		// GOOBS not available, connect.
		OBSMtx.Lock()

		var err error
		// goobs.WithLogger(config.Logger) (need a secondary logger to not pollute everything)
		OBS, err = goobs.New(config.RuntimeConfig.AV.OBS.Hostname, goobs.WithPassword(config.RuntimeConfig.AV.OBS.Password), goobs.WithLogger(log.Logger))
		// Not deferred as we need this back immediately.
		OBSMtx.Unlock()
		if err != nil {
			return err
		}
		if GoobsIsConnected() {
			log.Info("OBS connected.")
		}
	}
	return nil
}

func goobsDisconnect() (err error) {
	if OBS != nil {
		OBSMtx.Lock()
		defer OBSMtx.Unlock()
		err := OBS.Disconnect()
		if err != nil {
			return err
		}
		OBS = nil
	}
	return nil
}

func GoobsIsConnected() bool {
	OBSMtx.RLock()
	defer OBSMtx.RUnlock()
	if OBS == nil {
		return false
	}
	_, err := OBS.General.GetVersion()
	if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		return false
	}
	return err == nil
}

// OBSDaemon is responsible for watching the GOOBS connection on a regular basis and
// re-connecting if it seems to be unavailable for any reason.
func OBSDaemon(ctx context.Context) {
	err := goobsConnect()
	if err != nil {
		log.Errorf("Error connecting to OBS: %s", err)
	}
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if !GoobsIsConnected() {
				log.Infof("OBS disconnected, attempting to reconnect...")
				err = goobsConnect()
				if err != nil {
					log.Errorf("Failed to reconnect: %v", err)
				} else {
					log.Infof("Reconnected to OBS successfully")
				}
			}
		case <-ctx.Done():
			if OBS != nil {
				err := goobsDisconnect()
				if err != nil {
					log.Errorf("Error during disconnect: %v", err)
				}
				log.Infof("OBS disconnected gracefully.")
			}
			return
		}
	}
}
