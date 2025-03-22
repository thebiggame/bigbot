package ngtbg

// NodeCGDataboardReplicantData represents the NodeCG replicant state of the data shown in the Databoard.
type NodeCGDataboardReplicantData struct {
	EventNow  string `json:"now"`
	EventNext string `json:"next"`
}

type NodeCGMessageAlert struct {
	// The body of the alert that will be shown in the centre of the screen.
	Name string `json:"name"`
	// Whether the alert should arrive with "flair" - that is, an audible warning and bright text.
	Flair bool `json:"flair"`
}

const (
	// NodeCG Replicants

	// The current scheduled event. String.
	NodeCGReplicantScheduleNow = "schedule:now"
	// The next scheduled event. String.
	NodeCGReplicantScheduleNext = "schedule:next"

	// Whether the "event info" data is active. Boolean.
	NodeCGReplicantEventInfoActive = "event:info:active"
	// The body of "event info". String.
	NodeCGReplicantEventInfoBody = "event:info:body"

	// NodeCG Message channels

	// Fire an "alert" message. Use NodeCGMessageAlert to construct.
	NodeCGMessageChannelAlert = "alert"

	// Fire an "alert-end" message to animate out the alert. nil.
	NodeCGMessageChannelAlertEnd = "alert-end"
)
