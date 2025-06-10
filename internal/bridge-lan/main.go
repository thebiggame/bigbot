package bridge_lan

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/thebiggame/bigbot/internal/avcomms"
	"github.com/thebiggame/bigbot/internal/config"
	"github.com/thebiggame/bigbot/internal/log"
	protodef "github.com/thebiggame/bigbot/proto"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// logger stores the module's logger instance.
var logger = slog.New(slog.NewTextHandler(os.Stdout, nil)).With(slog.String("module", "bridge_lan"))

type BridgeLAN struct {
	// The context given to us by the main bot.
	ctx *context.Context

	// The websocket connection.
	conn *websocket.Conn

	// The app config.
	config config.BridgeConfig

	// Whether we are properly connected and authenticated with the WS server.
	connected bool
}

func New(config *config.BridgeConfig) (bridge *BridgeLAN, err error) {
	bridge = &BridgeLAN{
		config: *config,
	}
	return bridge, nil
}

func (mod *BridgeLAN) SetLogger(log *slog.Logger) {
	logger = log
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
				Key: string(bridge.config.Key),
			},
		},
	}
	msg, err := proto.Marshal(event)
	if err != nil {
		logger.Error("marshalling error", slog.Any("error", err))
	}
	err = bridge.conn.WriteMessage(websocket.BinaryMessage, msg)
	if err != nil {
		logger.Error("write error", slog.Any("error", err))
	}
	return nil
}

func (bridge *BridgeLAN) Start(ctx context.Context) (err error) {
	bridge.ctx = &ctx
	logger.Info("Connecting to BIGbot", slog.String("address", bridge.config.WsAddress))
	bridge.conn, _, err = websocket.DefaultDialer.Dial(bridge.config.WsAddress, nil)
	if err != nil {
		logger.Log(ctx, log.LevelFatal, "problem dialling BIGbot", slog.Any("error", err))
		return err
	}
	defer bridge.conn.Close()

	g, bridgeCtx := errgroup.WithContext(*bridge.ctx)

	err = bridge.doAuth()
	if err != nil {
		logger.Log(ctx, log.LevelFatal, "problem authenticating with BIGbot", slog.Any("error", err))
		return err
	}
	g.Go(func() error {
		avcomms.SetLogger(logger.With("module", "avcomms"))
		err := avcomms.Init(bridge.config.AV.NodeCG.Hostname, string(bridge.config.AV.NodeCG.AuthenticationKey))
		if err != nil {
			return err
		}
		avcomms.SetHostname(bridge.config.AV.OBS.Hostname)
		avcomms.SetPassword(string(bridge.config.AV.OBS.Password))
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
					logger.Error("problem reading from bridge connection", slog.Any("error", err))
					return err
				}
				event := &protodef.ServerEvent{}
				err = proto.Unmarshal(message, event)
				if err != nil {
					logger.Error("unmarshaling error", slog.Any("error", err))
				}
				logger.Log(ctx, log.LevelTrace, "unmarshalled", slog.Any("data", event))

				switch ev := event.Event.(type) {
				case *protodef.ServerEvent_Welcome:
					{
						logger.Info("Connected to BIGbot.", slog.String("remote_version", ev.Welcome.GetVersion()))
						bridge.connected = true
					}
				case *protodef.ServerEvent_Ping:
					{
						logger.Info("ping", slog.Any("data", event))
						// handlePing(clientEvent.GetPing(), c)
					}
				case *protodef.ServerEvent_ConnTermination:
					{
						logger.Error("connection terminated by BIGbot", slog.String("reason", ev.ConnTermination.GetMessage()))
						// Attempt graceful closure.
						return bridge.conn.Close()
					}
				case *protodef.ServerEvent_Version:
					logger.Debug("Version received")
					verObs, verNcg, err := bridge.handleVersions()
					var sCode int32
					if err != nil {
						logger.Error("handleNodeCGReplicantGet error", slog.Any("error", err))
						sCode = 500
					}
					var errData string
					if err != nil {
						errData = err.Error()
					}

					response := &protodef.ClientEvent{
						Event: &protodef.ClientEvent_RpcResponse{
							RpcResponse: &protodef.RPCResponse{
								RequestId:    event.RequestId,
								StatusCode:   sCode,
								ErrorMessage: errData,
								Payload: &protodef.RPCResponse_Versions{
									Versions: &protodef.VersionsResponse{
										Obs: verObs,
										Ncg: verNcg,
									},
								},
							},
						},
					}
					msg, err := proto.Marshal(response)
					if err != nil {
						logger.Error("marshalling error", slog.Any("error", err))
					}
					err = bridge.conn.WriteMessage(websocket.BinaryMessage, msg)
					if err != nil {
						logger.Error("write error", slog.Any("error", err))
					}
				case *protodef.ServerEvent_NodecgMessage:
					{
						logger.Debug("NodeCGMessage received")
						err := protoResponse(bridge.conn, event.RequestId, bridge.handleNodeCGMessageSend(ev))
						if err != nil {
							logger.Error("NodeCGMessageSend error", slog.Any("error", err))
						}
					}
				case *protodef.ServerEvent_NodecgReplicantSet:
					{
						logger.Debug("NodeCGReplicantSet received")
						err := protoResponse(bridge.conn, event.RequestId, bridge.handleNodeCGReplicantSet(ev))
						if err != nil {
							logger.Error("NodeCGReplicantSet error", slog.Any("error", err))
						}
					}
				case *protodef.ServerEvent_NodecgReplicantGet:
					{
						logger.Debug("NodeCGReplicantGet received")
						data, err := bridge.handleNodeCGReplicantGet(ev)
						var sCode int32
						if err != nil {
							logger.Error("handleNodeCGReplicantGet error", slog.Any("error", err))
							sCode = 500
						}
						var errData string
						if err != nil {
							errData = err.Error()
						}

						response := &protodef.ClientEvent{
							Event: &protodef.ClientEvent_RpcResponse{
								RpcResponse: &protodef.RPCResponse{
									RequestId:    event.RequestId,
									StatusCode:   sCode,
									ErrorMessage: errData,
									Payload: &protodef.RPCResponse_NcgReplicantGet{
										NcgReplicantGet: &protodef.NodecgReplicantGetResponse{
											Replicant: data,
										},
									},
								},
							},
						}
						msg, err := proto.Marshal(response)
						if err != nil {
							logger.Error("marshalling error", slog.Any("error", err))
						}
						err = bridge.conn.WriteMessage(websocket.BinaryMessage, msg)
						if err != nil {
							logger.Error("write error", slog.Any("error", err))
						}
					}
				case *protodef.ServerEvent_ObsSceneTransition:
					{
						logger.Debug("ObsSceneTransition received")
						err := protoResponse(bridge.conn, event.RequestId, bridge.handleOBSSceneTransition(ev))
						if err != nil {
							logger.Error("ObsSceneTransition error", slog.Any("error", err))
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
					logger.Error("write close error", slog.Any("error", err))
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
	if err := g.Wait(); err == nil || errors.Is(err, context.Canceled) || websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		logger.Info("Bridge stopped gracefully.")
	} else {
		logger.Warn("Error during shutdown", slog.Any("error", err))
	}
	return err
}
