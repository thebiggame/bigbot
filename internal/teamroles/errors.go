package teamroles

import (
	"errors"
	"fmt"
	"github.com/thebiggame/bigbot/internal/config"
)

var ErrAlreadyTeamMember = errors.New("You are already a member of that team")
var ErrNotTeamMember = errors.New("You are not a member of that team")
var ErrNotTeam = errors.New("This is not a team")
var ErrMaxTeamsReached = errors.New(fmt.Sprintf("You are already a member of %d or more teams! Please contact an administrator if you need more", config.RuntimeConfig.Teams.MaxUserTeams))
