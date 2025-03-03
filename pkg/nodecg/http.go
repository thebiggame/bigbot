package nodecg

import "net/http"

var HTTPClient http.Client

func init() {
	// Instantiate a default HTTP client.
	HTTPClient = http.Client{}
}
