package teamroles

import "testing"

var teamNameValidTests map[string]string = map[string]string{
	teamRolePrefix + " Test":               "Test",
	teamRolePrefix + " Test Æ":             "Test Æ",
	teamRolePrefix + " ✨":                  "✨",
	teamRolePrefix + "MergedForSomeReason": "MergedForSomeReason",
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
