package avbridge

// This file just stores constants / variables related to the OBS configuration state
// on the AVBOX running at the event.
// Please don't modify these during execution (they're variables because we need pointers,
// and you can't have a pointer to a constant.)

var (
	// Scenes
	sceneStandBy  = "Proj: Stand By"
	sceneDefault  = "Proj: Info Board (default)"
	sceneTestCard = "SPECIAL: Test Card"
	sceneBlack    = "SPECIAL: Black"

	// Transitions
	transitionCut     = "Cut"
	transitionFade    = "Fade"
	transitionStinger = "Stinger"
)

// ngtbgNodeCGDataboardReplicantData represents the NodeCG replicant state of the data shown in the Databoard.
type ngtbgNodeCGDataboardReplicantData struct {
	EventNow  string `json:"now"`
	EventNext string `json:"next"`
}

type ngtbgNodeCGMessageAlert struct {
	// The body of the alert that will be shown in the centre of the screen.
	Name string `json:"name"`
	// Whether the alert should arrive with "flair" - that is, an audible warning and bright text.
	Flair bool `json:"flair"`
}
