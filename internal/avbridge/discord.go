package avbridge

import "github.com/bwmarrin/discordgo"

var commands = []*discordgo.ApplicationCommand{
	{
		Name:                     "av",
		Description:              "📽️ Manage the AV setup (you must be a crew member)",
		DefaultMemberPermissions: &defaultAVCommandPermissions,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "status",
				Description: "📽️ Check current OBS link status.",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
			{
				Name:        "ftb",
				Description: "📽️ Fade Projector Output to Black.",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
		},
	},
}

func (mod *AVBridge) HandleDiscordCommand(s *discordgo.Session, i *discordgo.InteractionCreate) (handled bool, err error) {
	if i.ApplicationCommandData().Name != "av" {
		return false, nil
	}
	options := i.ApplicationCommandData().Options
	content := ""

	switch options[0].Name {
	case "status":
		if mod.goobsIsConnected() {
			content = "🙆 OBS is connected."
		} else {
			content = "🙅 OBS is **not connected.**"
		}
	case "ftb":
		content = "_Fading to Black..._"
	default:
		content = "😶 Please use a sub-command."
	}

	return true, s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

var defaultAVCommandPermissions int64 = discordgo.PermissionAdministrator
