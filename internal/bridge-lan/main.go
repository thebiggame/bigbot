package bridge_lan

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/thebiggame/bigbot/internal/avcomms"
	"github.com/thebiggame/bigbot/internal/log"
	protodef "github.com/thebiggame/bigbot/proto"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Config struct {
	WsAddress string `long:"addr" description:"BIGbot address and port" default:"ws://localhost:8080/ws" env:"ADDR"`
	Key       string `long:"key" description:"BIGbot authentication key" required:"" env:"KEY"`
	AV        struct {
		OBS struct {
			Hostname string `long:"host" help:"OBS Host" default:"" env:"HOST"`
			Password string `long:"password" help:"OBS password" default:"" env:"PASSWORD"`
		} `prefix:"obs." embed:"" envprefix:"OBS_"`
		NodeCG struct {
			Hostname          string `long:"host" help:"NodeCG Host" default:"" env:"HOST"`
			BundleName        string `long:"bundle" help:"NodeCG bundle name" default:"thebiggame" env:"BUNDLE"`
			AuthenticationKey string `long:"key" help:"Authentication key" default:"" env:"AUTHKEY"`
		} `prefix:"nodecg." embed:"" envprefix:"NODECG_"`
	} `prefix:"av." embed:"" envprefix:"AV_"`
}
type BridgeLAN struct {
	// The context given to us by the main bot.
	ctx *context.Context

	// The websocket connection.
	conn *websocket.Conn

	// The app config.
	config Config

	// Whether we are properly connected and authenticated with the WS server.
	connected bool
}

func New(config *Config) (bridge *BridgeLAN, err error) {
	bridge = &BridgeLAN{
		config: *config,
	}
	return bridge, nil
}

func (mod *BridgeLAN) Run() (err error) {
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

func (bridge *BridgeLAN) doAuth() error {
	event := &protodef.ClientEvent{
		Event: &protodef.ClientEvent_Authenticate{
			Authenticate: &protodef.Authenticate{
				Key: bridge.config.Key,
			},
		},
	}
	msg, err := proto.Marshal(event)
	if err != nil {
		log.Errorf("marshalling error: %s", err)
	}
	err = bridge.conn.WriteMessage(websocket.BinaryMessage, msg)
	if err != nil {
		log.Errorf("write error: %s", err)
	}
	return nil
}

func (bridge *BridgeLAN) handleNodeCGReplicantSet(event *protodef.ServerEvent_NodecgReplicantSet) error {
	err := avcomms.NodeCG.ReplicantSet(*bridge.ctx, event.NodecgReplicantSet.Namespace, event.NodecgReplicantSet.Replicant, event.NodecgReplicantSet.Data)
	if err != nil {
		return err
	}
	return nil
}

func (bridge *BridgeLAN) Start(ctx context.Context) (err error) {
	bridge.ctx = &ctx
	log.Infof("Connecting to BIGbot at %s...", bridge.config.WsAddress)
	bridge.conn, _, err = websocket.DefaultDialer.Dial(bridge.config.WsAddress, nil)
	if err != nil {
		log.Fatal("dial:", err)
		return err
	}
	defer bridge.conn.Close()

	g, bridgeCtx := errgroup.WithContext(*bridge.ctx)

	err = bridge.doAuth()
	if err != nil {
		log.Fatal("auth:", err)
		return err
	}
	g.Go(func() error {
		err := avcomms.Init(bridge.config.AV.NodeCG.Hostname, bridge.config.AV.NodeCG.AuthenticationKey)
		if err != nil {
			return err
		}
		avcomms.OBSDaemon(bridgeCtx)
		return nil
	})

	g.Go(func() error {
		for {
			select {
			case <-bridgeCtx.Done():
				return bridgeCtx.Err()
			default:
				_, message, err := bridge.conn.ReadMessage()
				if err != nil {
					log.Error("read:", err)
					return err
				}
				event := &protodef.ServerEvent{}
				err = proto.Unmarshal(message, event)
				if err != nil {
					log.Errorf("unmarshaling error: %s", err)
				}
				log.Tracef("unmarshalled: %s", event)

				switch ev := event.Event.(type) {
				case *protodef.ServerEvent_Welcome:
					{
						log.Infof("Welcome from BIGbot version %s", ev.Welcome.GetVersion())
						bridge.connected = true
					}
				case *protodef.ServerEvent_Ping:
					{
						log.Info("ping: %s", ev)
						// handlePing(clientEvent.GetPing(), c)
					}
				case *protodef.ServerEvent_NodecgReplicantSet:
					{
						log.Debug("NodeCGReplicantSet received: %s", ev)
						err := bridge.handleNodeCGReplicantSet(ev)
						var sCode int32
						if err != nil {
							log.Errorf("handleNodeCGReplicantSet error: %s", err)
							sCode = 500
						}

						response := &protodef.ClientEvent{
							Event: &protodef.ClientEvent_RpcResponse{
								RpcResponse: &protodef.RPCResponse{
									RequestId:    event.RequestId,
									StatusCode:   sCode,
									ErrorMessage: err.Error(),
								},
							},
						}
						msg, err := proto.Marshal(response)
						if err != nil {
							log.Errorf("marshalling error: %s", err)
						}
						err = bridge.conn.WriteMessage(websocket.BinaryMessage, msg)
						if err != nil {
							log.Errorf("write error: %s", err)
						}
					}
				}
			}
		}
	})

	// Goroutine for sending messages
	g.Go(func() error {
		for {
			select {
			case <-bridgeCtx.Done():
				// Do clean shutdown.
				// Cleanly close the connection by sending a close message and then
				// waiting (with timeout) for the server to close the connection.
				err = bridge.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Error("write close:", err)
					return err
				}
				select {
				case <-time.After(time.Second):
				}
				return bridgeCtx.Err()
			}
		}
	})

	// Closedown the context.
	if closeErr := g.Wait(); closeErr == nil || errors.Is(closeErr, context.Canceled) || websocket.IsCloseError(closeErr, websocket.CloseNormalClosure) {
		log.Info("Bridge stopped gracefully.")
	} else {
		log.Warn("Error during shutdown: %v", closeErr)
	}
	return
}
