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
