package teamroles

import "testing"

const (
	tnValidPrefix = "Team: "
)

var teamNameValidTests map[string]string = map[string]string{
	tnValidPrefix + "Test":   "Test",
	tnValidPrefix + "Test Æ": "Test Æ",
	tnValidPrefix + "✨":      "✨",
}

var teamNameInvalidTests []string = []string{
	"@everyone",
	"Test Team",
}

func TestGetTeamName(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		for team, name := range teamNameValidTests {
			valid, parsedName := getTeamName(team)
			if !valid {
				t.Errorf("Team Name '%s' should be valid", name)
			}
			if name != parsedName {
				t.Errorf("Team Name '%s' should be '%s'", name, parsedName)
			}
		}
	})
	t.Run("Invalid", func(t *testing.T) {
		for _, team := range teamNameInvalidTests {
			valid, parsedName := getTeamName(team)
			if valid {
				t.Errorf("Team Name '%s' parses as '%s' but should be invalid", team, parsedName)
			}
		}
	})
}
