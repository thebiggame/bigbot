package bridge_lan

import (
	"encoding/json"
	"errors"
	"github.com/andreykaipov/goobs/api/requests/scenes"
	"github.com/andreykaipov/goobs/api/requests/transitions"
	"github.com/thebiggame/bigbot/internal/avcomms"
	"github.com/thebiggame/bigbot/proto"
)

func (bridge *BridgeLAN) handleNodeCGMessageSend(event *proto.ServerEvent_NodecgMessage) error {
	err := avcomms.NodeCG.MessageSend(*bridge.ctx, event.NodecgMessage.Namespace, event.NodecgMessage.Channel, event.NodecgMessage.Data)
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

func (bridge *BridgeLAN) handleVersions() (obs, ncg string, err error) {
	// Get OBS version (and validate connection)
	if !avcomms.GoobsIsConnected() {
		return "", "", errors.New("OBS not connected")
	}
	verObs, err := avcomms.OBS.General.GetVersion()
	if err != nil {
		return "", "", err
	}

	// Get NodeCG version (and validate connection)
	repData, err := avcomms.NodeCG.ReplicantGetString(*bridge.ctx, "thebiggame", "bundle:version")
	if err != nil {
		return "", "", err
	}

	return verObs.ObsVersion, repData, nil
}

func (bridge *BridgeLAN) handleOBSSceneTransition(event *proto.ServerEvent_ObsSceneTransition) (err error) {
	if !avcomms.GoobsIsConnected() {
		return errors.New("OBS not connected")
	}
	// Set preview scene to the target
	_, err = avcomms.OBS.Scenes.SetCurrentPreviewScene(&scenes.SetCurrentPreviewSceneParams{
		SceneName: &event.ObsSceneTransition.SceneTarget,
	})
	if err != nil {
		return err
	}

	// set the desired transition
	_, err = avcomms.OBS.Transitions.SetCurrentSceneTransition(&transitions.SetCurrentSceneTransitionParams{
		TransitionName: &event.ObsSceneTransition.Transition,
	})
	if err != nil {
		return err
	}

	// then perform the transition
	_, err = avcomms.OBS.Transitions.TriggerStudioModeTransition(&transitions.TriggerStudioModeTransitionParams{})

	return err
}
