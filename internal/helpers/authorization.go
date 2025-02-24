package helpers

import (
	"github.com/bwmarrin/discordgo"
)

func UserIsCrew(s *discordgo.Session, guild string, u *discordgo.User) (isCrew bool, err error) {
	member, err := s.GuildMember(guild, u.ID)
	if err != nil {
		return false, err
	}
	for _, v := range member.Roles {
		role, err := s.State.Role(guild, v)
		if err != nil {
			return false, err
		}
		if role.Name == "crew" {
			return true, nil
		}
	}
	// User not a member of the Crew guild
	return false, nil
}
