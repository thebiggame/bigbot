package teamroles

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/thebiggame/bigbot/internal/config"
	"github.com/thebiggame/bigbot/internal/helpers"
	"regexp"
	"strings"
)

var Commands = []*discordgo.ApplicationCommand{
	{
		Name:        "team",
		Description: "Manage your membership of LAN teams",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "new",
				Description: "Create a new Team.",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "team-name",
						Description: "Name of the team you wish to create.",
						Required:    true,
					},
				},
			},
			{
				Name:        "join",
				Description: "Join an existing team",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionRole,
						Name:        "team",
						Description: "The team you wish to join.",
						Required:    true,
					},
				},
			},
			{
				Name:        "leave",
				Description: "Leave a team",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionRole,
						Name:        "team",
						Description: "The Team you wish to leave.",
						Required:    true,
					},
				},
			},
		},
	},
}

var CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"team": handleTeamCommand,
}

func handleTeamCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	content := ""

	// As you can see, names of subcommands (nested, top-level)
	// and subcommand groups are provided through the arguments.
	switch options[0].Name {

	case "new":
		// OOB check
		if len(options[0].Options) < 0 {
			content = "🤔 Please provide a team name."
			break
		}
		if i.Interaction.GuildID == "" {
			content = "😡 This command can only be used in a server."
			break
		}
		roleName := options[0].Options[0].StringValue()
		err := validateUserCanJoinRoleByName(s, i.Interaction.Member.User, i.GuildID, roleName)
		if err != nil {
			content = err.Error()
			break
		}
		role, err := createOrReturnRole(s, i.GuildID, roleName)
		if err != nil {
			content = err.Error()
			break
		}
		err = s.GuildMemberRoleAdd(i.GuildID, i.Interaction.Member.User.ID, role.ID)
		if err != nil {
			content = err.Error()
			break
		}
		content = fmt.Sprintln("🙌 Joined", role.Name)
	case "join":
		// OOB check
	case "leave":
		// OOB check
		if len(options[0].Options) < 0 {
			content = "🤔 Please provide a team name."
			break
		}
		if i.Interaction.GuildID == "" {
			content = "😡 This command can only be used in a server."
			break
		}
		role := options[0].Options[0].RoleValue(s, i.GuildID)
		isTeam := validateRoleIsTeam(role.Name)
		if !isTeam {
			content = fmt.Sprintf("⚠️ %s", ErrNotTeam)
			break
		}
		err := validateUserIsRoleMember(s, i.Interaction.Member.User, i.GuildID, role)
		if err != nil {
			content = fmt.Sprintf("⚠️ %s", err.Error())
			break
		}
		err = s.GuildMemberRoleRemove(i.GuildID, i.Interaction.Member.User.ID, role.ID)
		if err != nil {
			content = fmt.Sprintf("⚠️ %s", err.Error())
			break
		}
		content = fmt.Sprintln("👋 Left", role.Name)
	default:
		content = "😶 Please use a sub-command."
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func validateUserCanJoinRoleByName(s *discordgo.Session, u *discordgo.User, guild, targetRole string) error {
	// This function validates that the given GuildMember satisfies the following rules:
	// - is not already assigned to more than 5 Team Roles
	// - is not already assigned to the given targetRole
	var roleCount int
	member, err := s.GuildMember(guild, u.ID)
	if err != nil {
		return err
	}
	for _, v := range member.Roles {
		role, err := s.State.Role(guild, v)
		if err != nil {
			return err
		}
		// TODO This whole thing of matching on name sucks. Refactor me.
		getRoleName := regexp.MustCompile(`(?i)^(?:Team):* ?(.*)`)
		roleName := getRoleName.FindAllStringSubmatch(role.Name, -1)
		// role names get normalized to lower case during the lookup only
		if roleName != nil && strings.ToLower(roleName[0][1]) == strings.ToLower(targetRole) {
			// The Member is already part of the given GuildRole!
			return ErrAlreadyTeamMember
		}
		// Check if it's a team role, and if it is, add to the counter
		if strings.HasPrefix(role.Name, "Team:") {
			roleCount += 1
		}
	}
	if roleCount >= config.RuntimeConfig.MaxUserRoles {
		// Joining this Role would take the user over their limit
		return ErrMaxTeamsReached
	}

	// Succ(ess)
	return nil
}

func validateUserCanJoinRole(s *discordgo.Session, u *discordgo.User, guild string, targetRole *discordgo.Role) (err error) {
	// TODO This whole thing of matching on name sucks. Refactor me.
	roleName := targetRole.Name
	return validateUserCanJoinRoleByName(s, u, guild, roleName)
}

func validateUserIsRoleMember(s *discordgo.Session, u *discordgo.User, guild string, targetRole *discordgo.Role) error {
	// This function validates that the given GuildMember satisfies the following rules:
	// - is assigned to the given guild role
	member, err := s.GuildMember(guild, u.ID)
	if err != nil {
		return err
	}
	for _, v := range member.Roles {
		role, err := s.State.Role(guild, v)
		if err != nil {
			return err
		}
		if role != nil && role.ID == targetRole.ID {
			return nil
		}
	}

	// Failure
	return ErrNotTeamMember
}

func validateRoleIsTeam(roleName string) (isTeam bool) {
	getRoleName := regexp.MustCompile(`(?i)^(?:Team):* ?(.*)`)
	return getRoleName.MatchString(roleName)
}

func createOrReturnRole(s *discordgo.Session, guild string, rname string) (v *discordgo.Role, err error) {
	roles, err := s.GuildRoles(guild)
	getRole := regexp.MustCompile(`(?i)^(?:Team):*`)
	if !getRole.MatchString(rname) {
		rname = fmt.Sprintln("Team:", rname)
	}
	rname = strings.Replace(rname, "\n", "", -1)
	if err == nil {
		for _, v := range roles {
			// role names get normalized to lower case during the lookup only
			if strings.ToLower(v.Name) == strings.ToLower(rname) {
				helpers.LogDebug("tying", rname, "to old role", v.Name)
				return v, nil
			}
		}
		// couldn't find the role in our list, create it
		helpers.LogDebug("tying", rname, "to new role", rname)
		var rColour = 8290694
		var rHoist = true
		var rMentionable = true
		var rPerms int64 = 0
		rParams := discordgo.RoleParams{
			Name:         rname,
			Color:        &rColour,
			Hoist:        &rHoist,
			Permissions:  &rPerms,
			Mentionable:  &rMentionable,
			UnicodeEmoji: nil,
			Icon:         nil,
		}
		role, err := s.GuildRoleCreate(guild, &rParams)
		return role, err
	}
	return nil, fmt.Errorf("problem creating the target role: %w", err)
}
