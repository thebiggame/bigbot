package ngtbg

// This file just stores constants / variables related to the OBS configuration state
// on the AVBOX running at the event.
// Please don't modify these during execution (they're variables because we need pointers,
// and you can't have a pointer to a constant.)

var (
	// Scenes
	OBSSceneStandBy  = "Proj: Stand By"
	OBSSceneDefault  = "Proj: Info Board (default)"
	OBSSceneTestCard = "SPECIAL: Test Card"
	OBSSceneBlack    = "SPECIAL: Black"

	// Transitions
	OBSTransCut             = "Cut"
	OBSTransFade            = "Fade"
	OBSTransStingModernWipe = "tBG Modern Wipe"
)
