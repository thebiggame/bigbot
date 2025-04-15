package teamroles

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/thebiggame/bigbot/internal/config"
	"github.com/thebiggame/bigbot/internal/helpers"
	"github.com/thebiggame/bigbot/internal/log"
	"regexp"
	"strings"
)

const (
	teamRolePrefix = "Team:"
)

type TeamRoles struct {
	discord *discordgo.Session
	close   chan bool
}

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "team",
		Description: "🧑‍🤝‍🧑 Manage your membership of LAN teams.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "new",
				Description: "🧑‍🤝‍🧑✨ Create a new Team.",
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
				Description: "🧑‍🤝‍🧑🤝 Join an existing team",
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
				Description: "🧑‍🤝‍🧑👋 Leave a team",
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

func New(discord *discordgo.Session) *TeamRoles {
	return &TeamRoles{
		discord: discord,
	}
}

func (mod *TeamRoles) DiscordCommands() ([]*discordgo.ApplicationCommand, error) {
	return commands, nil
}

func (mod *TeamRoles) Start(ctx context.Context) (err error) {
	// This module simply registers handlers, and does not need to run continuously (so we don't need the context)
	return ctx.Err()
}

func (mod *TeamRoles) DiscordHandleMessage(session *discordgo.Session, message *discordgo.MessageCreate) (err error) {
	return nil
}

func (mod *TeamRoles) DiscordHandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) (handled bool, err error) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		if i.ApplicationCommandData().Name != "team" {
			return false, nil
		}

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
				content = fmt.Sprintf("⚠️ %s", err.Error())
				break
			}
			role, exists, err := createOrReturnRole(s, i.GuildID, roleName)
			if err != nil {
				content = fmt.Sprintf("⚠️ %s", err.Error())
				break
			}
			err = s.GuildMemberRoleAdd(i.GuildID, i.Interaction.Member.User.ID, role.ID)
			if err != nil {
				content = fmt.Sprintf("⚠️ %s", err.Error())
				break
			}
			if exists {
				content = fmt.Sprintln("🤝 Joined existing", role.Name)
			} else {
				content = fmt.Sprintln("✨ Created", role.Name)
			}

		case "join":
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
			isTeam, _ := getTeamName(role.Name)
			if !isTeam {
				content = fmt.Sprintf("⚠️ %s. Stop that. <:ninja:449495170430533633>", ErrNotTeam)
				break
			}
			err := validateUserCanJoinRole(s, i.Interaction.Member.User, i.GuildID, role)
			if err != nil {
				content = fmt.Sprintf("⚠️ %s", err.Error())
				break
			}
			err = s.GuildMemberRoleAdd(i.GuildID, i.Interaction.Member.User.ID, role.ID)
			if err != nil {
				content = fmt.Sprintf("⚠️ %s", err.Error())
				break
			}
			content = fmt.Sprintln("🤝 Joined", role.Name)
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
			isTeam, _ := getTeamName(role.Name)
			if !isTeam {
				content = fmt.Sprintf("⚠️ %s. Stop that. <:ninja:449495170430533633>", ErrNotTeam)
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

		return true, helpers.DiscordInteractionEphemeralResponse(s, i, content)
	default:
		return false, nil
	}
}

func validateUserCanJoinRoleByName(s *discordgo.Session, u *discordgo.User, guild, targetRole string) error {
	// This function validates that the given GuildMember satisfies the following rules:
	// - is not already assigned to more than 5 Team Roles
	// - is trying to join a team
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
		isTeam, _ := getTeamName(role.Name)
		// role names get normalized to lower case during the lookup only
		if isTeam {
			if strings.ToLower(role.Name) == strings.ToLower(targetRole) {
				// The Member is already part of the given GuildRole!
				return ErrAlreadyTeamMember
			}
			// Check if it's a team role, and if it is, add to the counter
			roleCount++
		}
	}
	if roleCount >= config.RuntimeConfig.Teams.MaxUserTeams {
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

func getTeamName(roleName string) (isTeam bool, team string) {
	getRoleName := regexp.MustCompile(`(?i)^` + teamRolePrefix + `* ?(.*)`)
	teamName := getRoleName.FindAllStringSubmatch(roleName, -1)
	if teamName != nil && teamName[0][1] != "" {
		return true, teamName[0][1]
	}
	return false, ""
}

func createOrReturnRole(s *discordgo.Session, guild string, rname string) (v *discordgo.Role, roleExisted bool, err error) {
	roles, err := s.GuildRoles(guild)
	isRole, _ := getTeamName(rname)
	if !isRole {
		rname = fmt.Sprintln(teamRolePrefix, rname)
	}
	rname = strings.Replace(rname, "\n", "", -1)
	if err == nil {
		for _, v := range roles {
			// role names get normalized to lower case during the lookup only
			if strings.ToLower(v.Name) == strings.ToLower(rname) {
				log.Debugf("Tying %s to existing team role %s", rname, v.Name)
				return v, true, nil
			}
		}
		// couldn't find the role in our list, create it
		log.Debugf("Creating new team role %s", rname)
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
		return role, false, err
	}
	return nil, false, fmt.Errorf("problem creating the target role: %w", err)
}
