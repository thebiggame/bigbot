package bridge_wan

import (
	"errors"
	"github.com/gorilla/websocket"
	"github.com/thebiggame/bigbot/internal/log"
	protodef "github.com/thebiggame/bigbot/proto"
	"google.golang.org/protobuf/proto"
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
		log.Errorf("marshalling error: %s", err)
	}
	err = c.WriteMessage(websocket.BinaryMessage, msg)
	if err != nil {
		log.Errorf("write error: %s", err)
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
		log.Errorf("marshalling error: %s", err)
	}
	err = c.WriteMessage(websocket.BinaryMessage, msg)
	if err != nil {
		log.Errorf("write error: %s", err)
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
			log.Error("write close:", err)
		}
	}

	log.Infof("BIGbridge connected from %s", c.RemoteAddr())

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
		log.Debug("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Debug("read:", err)
			break
		}
		log.Tracef("received: %s", message)
		clientEvent := &protodef.ClientEvent{}
		err = proto.Unmarshal(message, clientEvent)
		if err != nil {
			log.Errorf("unmarshaling error: %s", err)
			continue
		}
		log.Tracef("unmarshalled: %s", clientEvent)

		switch event := clientEvent.Event.(type) {
		case *protodef.ClientEvent_Ping:
			{
				handlePing(event.Ping, c)
			}
		case *protodef.ClientEvent_Authenticate:
			{
				err := bridge.handleAuthenticate(event.Authenticate, c)
				if err != nil {
					log.Errorf("authentication error: %s", err)
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
				log.Warnf("No matching request for RPC response with ID %s", event.RpcResponse.RequestId)
			}
			bridge.wsResponseMtx.Unlock()
		}

		if clientEvent.GetPing() != nil {
			handlePing(clientEvent.GetPing(), c)
		}

	}
	log.Debug("Ended a client session.")
}
