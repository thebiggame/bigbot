package bridge_wan

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/thebiggame/bigbot/internal/config"
	"github.com/thebiggame/bigbot/internal/log"
	protodef "github.com/thebiggame/bigbot/proto"
	"google.golang.org/protobuf/proto"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type BridgeWAN struct {
	// The context given to us by the main bot.
	ctx *context.Context

	// The HTTP server.
	httpServer *http.Server

	// The authentication key to test WS connections against.
	wsKey string

	// Active wvent websocket connection.
	wsConn *websocket.Conn

	wsResponseCh  map[string]chan *protodef.RPCResponse
	wsResponseMtx sync.Mutex
}

func generateRequestID() string {
	// Generate a unique ID (use UUIDs in production)
	return time.Now().Format("20060102150405.000000000")
}

var EventBridge *BridgeWAN

func BridgeIsAvailable() (up bool) {
	if EventBridge != nil {
		return EventBridge.EventAvailable()
	}
	return false
}

func (bridge *BridgeWAN) BrReplicantSet(bundle, replicant string, value interface{}) (err error) {
	if EventBridge == nil {
		return errors.New("EventBridge not initialised")
	}
	if EventBridge.wsConn == nil {
		return errors.New("EventBridge not connected")
	}

	// Get an idempotency key for this request
	requestID := generateRequestID()
	// Create a channel to receive the response
	responseCh := make(chan *protodef.RPCResponse, 1)

	// Store the channel in the responseCh map
	bridge.wsResponseMtx.Lock()
	bridge.wsResponseCh[requestID] = responseCh
	bridge.wsResponseMtx.Unlock()

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	event := &protodef.ServerEvent{
		RequestId: requestID,
		Event: &protodef.ServerEvent_NodecgReplicantSet{
			NodecgReplicantSet: &protodef.NodecgReplicantSet{
				Namespace: bundle,
				Replicant: replicant,
				Data:      data,
			},
		},
	}
	msg, err := proto.Marshal(event)
	if err != nil {
		return err
	}
	err = EventBridge.wsConn.WriteMessage(websocket.BinaryMessage, msg)
	if err != nil {
		return err
	}

	// Wait for the response or timeout
	select {
	case rpcResponse := <-responseCh:
		// Handle server-side errors
		if rpcResponse.StatusCode != 0 {
			return errors.New(rpcResponse.ErrorMessage)
		}

		// Deserialize the response payload into the provided response message
		return nil
	case <-time.After(time.Second * 10):
		// Clean up the channel on timeout
		bridge.wsResponseMtx.Lock()
		delete(bridge.wsResponseCh, requestID)
		bridge.wsResponseMtx.Unlock()
		return errors.New("RPC call timed out")
	}
}

func New() (bridge *BridgeWAN, err error) {
	bridge = &BridgeWAN{
		httpServer:   &http.Server{Addr: config.RuntimeConfig.Bridge.Address},
		wsKey:        config.RuntimeConfig.Bridge.Key,
		wsResponseCh: make(map[string]chan *protodef.RPCResponse),
	}
	EventBridge = bridge
	return bridge, nil
}

func (mod *BridgeWAN) Run() (err error) {
	// Create app context (this is passed to modules).
	// The signal.NotifyContext is a special context that gets torn down when an interrupt / SIGTERM is received.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err = mod.Start(ctx)

	if err != nil {
		return err
	}
	return
}

func (mod *BridgeWAN) Start(ctx context.Context) (err error) {
	if !config.RuntimeConfig.Bridge.Enabled {
		return nil
	}

	mod.ctx = &ctx

	http.HandleFunc("/ws", mod.wsHandle)
	go func() {
		log.Infof("BIGbridge SERVER listening at %v", mod.httpServer.Addr)
		if err := mod.httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("error with HTTP server: %v", err)
		}
	}()

	for {
		select {
		// Spinloop here to make sure that we stay alive long enough for the context to get torn down properly.
		case <-ctx.Done():
			// Do clean shutdown.
			if err := mod.httpServer.Shutdown(context.Background()); err != nil {
				log.Errorf("error shutting down http server: %v", err)
			}
			return ctx.Err()
		}
	}
}
