package ngtbg

type NodeCGReplicantDataAlertData struct {
	// The body of the alert that will be shown in the centre of the screen.
	Body string `json:"body"`
	// Whether the alert should arrive with "flair" - that is, an audible warning and bright text.
	Flair bool `json:"flair"`
	// How long to wait before showing the alert animation.
	Delay int `json:"delay"`
}

// NodeCGReplicantDataMusicData is the content of a MusicData replicant.
type NodeCGReplicantDataMusicData struct {
	// The song title.
	Title string `json:"title"`
	// The song artist.
	Artist string `json:"artist"`
}

// NodeCGReplicantDataShoutboxEntry is equivalent to one Shout.
type NodeCGReplicantDataShoutboxEntry struct {
	// A unique ID for this message.
	ID   string `json:"id"`
	User struct {
		// The name of the user who submitted this Shout.
		Name string `json:"name"`
		// The URL that their avatar can be found at.
		Avatar string `json:"avatar_url"`
	} `json:"user"`
	// The timestamp of this shout.
	Timestamp string `json:"timestamp"`
	// The shout message content (in markdown).
	Message string `json:"message"`
}

type NodeCGReplicantDataShoutboxEntries struct {
	Shouts []NodeCGReplicantDataShoutboxEntry `json:"shouts"`
}

const (
	// NodeCG Replicants

	// Whether the "event info" data is active. Boolean.
	NodeCGReplicantEventInfoActive = "event:info:active"
	// The body of "event info". String.
	NodeCGReplicantEventInfoBody = "event:info:body"

	// Information on the current music traka. Object of type NodeCGReplicantDataMusicData
	NodeCGReplicantMusicData = "music:data"

	// Whether the notification "alert" is active. Boolean.
	NodeCGReplicantNotificationAlertActive = "notify:alert:active"

	// The data for the "alert" notification type. Object of type NodeCGReplicantDataAlertData
	NodeCGReplicantNotificationAlertData = "notify:alert:data"

	// A list of "shouts". Object of type NodeCGReplicantDataShoutboxEntries
	NodeCGReplicantShoutbox = "shoutbox:messages"

	// NodeCG Message channels

	// Fire an "alert" message. Use NodeCGMessageAlert to construct.
	// Deprecated, removed in v53.2
	// NodeCGMessageChannelAlert = "alert"

	// Fire an "alert-end" message to animate out the alert. nil.
	// Deprecated, removed in v53.2
	//NodeCGMessageChannelAlertEnd = "alert-end"

	// Fire an "shoutbox:message-new" message with the contents of a new Shout. Use NodeCGMessageShout to construct.
	NodeCGMessageShoutboxNew = "shoutbox:new-discord"
)
