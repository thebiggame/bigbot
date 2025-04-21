package bridge_lan

import (
	"encoding/json"
	"github.com/thebiggame/bigbot/internal/avcomms"
	"github.com/thebiggame/bigbot/proto"
)

func (bridge *BridgeLAN) handleNodeCGMessageSend(event *proto.ServerEvent_NodecgMessage) error {
	err := avcomms.NodeCG.ReplicantSet(*bridge.ctx, event.NodecgMessage.Namespace, event.NodecgMessage.Channel, event.NodecgMessage.Data)
	if err != nil {
		return err
	}
	return nil
}

func (bridge *BridgeLAN) handleNodeCGReplicantSet(event *proto.ServerEvent_NodecgReplicantSet) error {
	err := avcomms.NodeCG.ReplicantSet(*bridge.ctx, event.NodecgReplicantSet.Namespace, event.NodecgReplicantSet.Replicant, event.NodecgReplicantSet.Data)
	if err != nil {
		return err
	}
	return nil
}

func (bridge *BridgeLAN) handleNodeCGReplicantGet(event *proto.ServerEvent_NodecgReplicantGet) (data []byte, err error) {
	repData, err := avcomms.NodeCG.ReplicantGet(*bridge.ctx, event.NodecgReplicantGet.GetNamespace(), event.NodecgReplicantGet.GetReplicant())
	if err != nil {
		return nil, err
	}
	data, err = json.Marshal(repData)
	return data, err
}
