package bridge_lan

import (
	"github.com/gorilla/websocket"
	protodef "github.com/thebiggame/bigbot/proto"
	"google.golang.org/protobuf/proto"
	"log/slog"
)

func protoResponse(bridgeConn *websocket.Conn, requestId string, err error) error {
	var sCode int32
	var errData string
	if err != nil {
		logger.Error("ObsSceneTransition error", slog.Any("error", err))
		sCode = 500
		errData = err.Error()
	}

	response := &protodef.ClientEvent{
		Event: &protodef.ClientEvent_RpcResponse{
			RpcResponse: &protodef.RPCResponse{
				RequestId:    requestId,
				StatusCode:   sCode,
				ErrorMessage: errData,
			},
		},
	}
	msg, respErr := proto.Marshal(response)
	if respErr != nil {
		logger.Error("marshalling error", slog.Any("error", err))
		return respErr
	}
	respErr = bridgeConn.WriteMessage(websocket.BinaryMessage, msg)
	if respErr != nil {
		logger.Error("write error", slog.Any("error", err))
		return respErr
	}
	return err
}
