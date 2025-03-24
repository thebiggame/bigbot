package helpers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/thebiggame/bigbot/internal/config"
	"github.com/thebiggame/bigbot/internal/log"
)

func UserIsCrew(s *discordgo.Session, guild string, u *discordgo.User) (isCrew bool, err error) {
	var crewRoleID = config.RuntimeConfig.Discord.Permissions.CrewRole
	if crewRoleID == "" {
		// Crew role not set, unable to parse.
		log.Warn("Crew lookup performed with crew Role ID not defined in config - returning false by default for safety. Please define a role ID!")
		return false, nil
	}
	member, err := s.GuildMember(guild, u.ID)
	if err != nil {
		return false, err
	}
	for _, v := range member.Roles {
		role, err := s.State.Role(guild, v)
		if err != nil {
			return false, err
		}
		if role.ID == crewRoleID {
			return true, nil
		}
	}
	// User not a member of the Crew guild
	return false, nil
}
