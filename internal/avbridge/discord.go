package avbridge

import (
	"github.com/bwmarrin/discordgo"
	"github.com/thebiggame/bigbot/internal/avbridge/ngtbg"
	"github.com/thebiggame/bigbot/internal/avcomms"
	bridge_wan "github.com/thebiggame/bigbot/internal/bridge-wan"
	"github.com/thebiggame/bigbot/internal/helpers"
)

var commands = []*discordgo.ApplicationCommand{
	{
		Name:                     "av",
		Description:              "📽️ Manage the AV setup (you must be a crew member)",
		DefaultMemberPermissions: &defaultAVCommandPermissions,
		DMPermission:             &defaultAVCommandDMPermissions,
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
			{
				Name:        "infoboard",
				Description: "📽️ Transition to Infoboard (the default projector display).",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
		},
	},
}

func (mod *AVBridge) DiscordHandleMessage(session *discordgo.Session, message *discordgo.MessageCreate) (err error) {
	return nil
}

func (mod *AVBridge) DiscordHandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) (handled bool, err error) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		// Handle normally.
		if i.ApplicationCommandData().Name != "av" {
			return false, nil
		}
		options := i.ApplicationCommandData().Options
		content := "😶 Unknown command..."

		switch options[0].Name {
		case "status":
			if avcomms.GoobsIsConnected() {
				content = "🙆 OBS is connected."
			} else {
				content = "🙅 OBS is **not connected.**"
			}
		case "ftb":
			if !avcomms.GoobsIsConnected() {
				content = "🙅 OBS is **not connected.**"
				break
			}
			// Let the client know we're working on it.
			if helpers.DiscordDeferEphemeralInteraction(s, i) != nil {
				return true, err
			}
			return true, mod.discordCommandAVFTB(s, i)
		case "infoboard":
			if !avcomms.GoobsIsConnected() {
				content = "🙅 OBS is **not connected.**"
				break
			}
			// Let the client know we're working on it.
			if helpers.DiscordDeferEphemeralInteraction(s, i) != nil {
				return true, err
			}
			return true, mod.discordCommandAVInfoboard(s, i)
		}

		// Not handled by specific handler function, respond with content data.
		return true, helpers.DiscordInteractionEphemeralResponse(s, i, content)
	default:
		// Not something we recognise.
		return false, nil
	}

}

func (mod *AVBridge) discordCommandAVFTB(s *discordgo.Session, i *discordgo.InteractionCreate) (err error) {
	err = bridge_wan.EventBridge.OBSSceneTransition(ngtbg.OBSSceneBlack, ngtbg.OBSTransFade)
	if err != nil {
		return err
	}

	// Finally, confirm we did the thing.
	_, err = helpers.DiscordInteractionFollowupMessage(s, i, "_Fading to black_...")
	return err
}

func (mod *AVBridge) discordCommandAVInfoboard(s *discordgo.Session, i *discordgo.InteractionCreate) (err error) {
	err = bridge_wan.EventBridge.OBSSceneTransition(ngtbg.OBSSceneDefault, ngtbg.OBSTransStingModernWipe)
	if err != nil {
		return err
	}

	// Finally, confirm we did the thing.
	_, err = helpers.DiscordInteractionFollowupMessage(s, i, "Switched to Infoboard.")
	return err
}

var defaultAVCommandPermissions int64 = discordgo.PermissionAdministrator
var defaultAVCommandDMPermissions = false
