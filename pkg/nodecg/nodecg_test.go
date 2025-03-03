package nodecg

import (
	"encoding/json"
	"testing"
)

func TestNodeCGParseReplicantObject(t *testing.T) {
	dataStr := "{\"status\":\"OK\",\"name\":\"data\",\"bundle\":\"thebiggame\",\"value\":{\"now\":\"The Next Generation of theBIGGAME AV.\",\"next\":\"tBG53\"}}"
	var resp replicantResponse
	err := json.Unmarshal([]byte(dataStr), &resp)
	if err != nil {
		t.Fatalf("Unmarshal err: %v", err)
	}
}

func TestNodeCGParseReplicantBool(t *testing.T) {
	dataStr := "{\"status\":\"OK\",\"name\":\"active\",\"bundle\":\"thebiggame\",\"value\":false}"
	var resp replicantResponse
	err := json.Unmarshal([]byte(dataStr), &resp)
	if err != nil {
		t.Fatalf("Unmarshal err: %v", err)
	}
}
