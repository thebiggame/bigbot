package bridge_wan

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/thebiggame/bigbot/proto"
	proto2 "google.golang.org/protobuf/proto"
	"time"
)

func (bridge *BridgeWAN) BrReplicantGet(bundle, replicant string, target any) (err error) {
	if EventBridge == nil {
		return errors.New("EventBridge not initialised")
	}
	if EventBridge.wsConn == nil {
		return errors.New("EventBridge not connected")
	}

	// Get an idempotency key for this request
	requestID := generateRequestID()
	// Create a channel to receive the response
	responseCh := make(chan *proto.RPCResponse, 1)

	// Store the channel in the responseCh map
	bridge.wsResponseMtx.Lock()
	bridge.wsResponseCh[requestID] = responseCh
	bridge.wsResponseMtx.Unlock()

	event := &proto.ServerEvent{
		RequestId: requestID,
		Event: &proto.ServerEvent_NodecgReplicantGet{
			NodecgReplicantGet: &proto.NodecgReplicantGet{
				Namespace: bundle,
				Replicant: replicant,
			},
		},
	}
	msg, err := proto2.Marshal(event)
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
		if rpcResponse.GetNcgReplicantGet() == nil {
			return errors.New("RPC response not of type NcgReplicantGet")
		}
		err = json.Unmarshal(rpcResponse.GetNcgReplicantGet().GetReplicant(), &target)

		return err
	case <-time.After(time.Second * 10):
		// Clean up the channel on timeout
		bridge.wsResponseMtx.Lock()
		delete(bridge.wsResponseCh, requestID)
		bridge.wsResponseMtx.Unlock()
		return errors.New("RPC call timed out")
	}
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
	responseCh := make(chan *proto.RPCResponse, 1)

	// Store the channel in the responseCh map
	bridge.wsResponseMtx.Lock()
	bridge.wsResponseCh[requestID] = responseCh
	bridge.wsResponseMtx.Unlock()

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	event := &proto.ServerEvent{
		RequestId: requestID,
		Event: &proto.ServerEvent_NodecgReplicantSet{
			NodecgReplicantSet: &proto.NodecgReplicantSet{
				Namespace: bundle,
				Replicant: replicant,
				Data:      data,
			},
		},
	}
	msg, err := proto2.Marshal(event)
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

func (bridge *BridgeWAN) BrMessageSend(bundle, channel string, value interface{}) (err error) {
	if EventBridge == nil {
		return errors.New("EventBridge not initialised")
	}
	if EventBridge.wsConn == nil {
		return errors.New("EventBridge not connected")
	}

	// Get an idempotency key for this request
	requestID := generateRequestID()
	// Create a channel to receive the response
	responseCh := make(chan *proto.RPCResponse, 1)

	// Store the channel in the responseCh map
	bridge.wsResponseMtx.Lock()
	bridge.wsResponseCh[requestID] = responseCh
	bridge.wsResponseMtx.Unlock()

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	event := &proto.ServerEvent{
		RequestId: requestID,
		Event: &proto.ServerEvent_NodecgMessage{
			NodecgMessage: &proto.NodecgMessageSend{
				Namespace: bundle,
				Channel:   channel,
				Data:      data,
			},
		},
	}
	msg, err := proto2.Marshal(event)
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
		return nil
	case <-time.After(time.Second * 10):
		// Clean up the channel on timeout
		bridge.wsResponseMtx.Lock()
		delete(bridge.wsResponseCh, requestID)
		bridge.wsResponseMtx.Unlock()
		return errors.New("RPC call timed out")
	}
}
