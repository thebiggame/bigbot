package bridge_wan

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/thebiggame/bigbot/internal/log"
	protodef "github.com/thebiggame/bigbot/proto"
	"google.golang.org/protobuf/proto"
	"log/slog"
	"net"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handlePing(m *protodef.Ping, c *websocket.Conn) {
	response := new(protodef.Ping)
	msg, err := proto.Marshal(response)
	if err != nil {
		logger.Error("marshalling error", slog.Any("error", err))
	}
	err = c.WriteMessage(websocket.BinaryMessage, msg)
	if err != nil {
		logger.Error("write error", slog.Any("error", err))
	}
}

func writeWelcome(c *websocket.Conn) {
	event := &protodef.ServerEvent{
		Event: &protodef.ServerEvent_Welcome{
			Welcome: &protodef.Welcome{
				Version: "v53.4.0",
			},
		},
	}
	msg, err := proto.Marshal(event)
	if err != nil {
		logger.Error("marshalling error", slog.Any("error", err))
	}
	err = c.WriteMessage(websocket.BinaryMessage, msg)
	if err != nil {
		logger.Error("write error", slog.Any("error", err))
	}
}

func (bridge *BridgeWAN) handleAuthenticate(m *protodef.Authenticate, c *websocket.Conn) (err error) {
	key := m.GetKey()
	if key != bridge.wsKey {
		// Authentication failed.
		return errors.New("invalid key")
	}
	// Boot any existing connection.
	if bridge.wsConn != nil {
		err = bridge.wsConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil && !errors.Is(err, net.ErrClosed) {
			logger.Error("write close", slog.Any("error", err))
		}
	}

	logger.Info("BIGbridge connected", slog.String("address", c.RemoteAddr().String()))

	// Set this connection as the valid connection.
	bridge.wsConn = c
	writeWelcome(c)
	return nil
}

func (bridge *BridgeWAN) EventAvailable() bool {
	if bridge.wsConn == nil {
		return false
	}
	return true
}

func (bridge *BridgeWAN) wsHandle(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Warn("upgrade", slog.Any("error", err))
		return
	}
	defer c.Close()
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			logger.Warn("failed to read message from wire", slog.Any("error", err))
			break
		}
		logger.Log(context.Background(), log.LevelTrace, "received: %s", message)
		clientEvent := &protodef.ClientEvent{}
		err = proto.Unmarshal(message, clientEvent)
		if err != nil {
			logger.Error("unmarshaling error", slog.Any("error", err))
			continue
		}
		logger.Log(context.Background(), log.LevelTrace, fmt.Sprintf("unmarshalled: %s", clientEvent))

		switch event := clientEvent.Event.(type) {
		case *protodef.ClientEvent_Ping:
			{
				handlePing(event.Ping, c)
			}
		case *protodef.ClientEvent_Authenticate:
			{
				err := bridge.handleAuthenticate(event.Authenticate, c)
				if err != nil {
					logger.Error("authentication error", slog.Any("error", err))
					return
				}
			}
		case *protodef.ClientEvent_RpcResponse:
			// It's a response to a previous request.
			bridge.wsResponseMtx.Lock()
			if ch, ok := bridge.wsResponseCh[event.RpcResponse.RequestId]; ok {
				ch <- event.RpcResponse
				close(ch)
				delete(bridge.wsResponseCh, event.RpcResponse.RequestId)
			} else {
				logger.Warn("No matching request for RPC response", slog.String("request_id", event.RpcResponse.RequestId))
			}
			bridge.wsResponseMtx.Unlock()
		}

		if clientEvent.GetPing() != nil {
			handlePing(clientEvent.GetPing(), c)
		}

	}
	logger.Info("Ended client session", slog.String("address", c.RemoteAddr().String()))
}
