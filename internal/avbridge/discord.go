package avbridge

import (
	"github.com/andreykaipov/goobs/api/requests/scenes"
	"github.com/andreykaipov/goobs/api/requests/transitions"
	"github.com/bwmarrin/discordgo"
	"github.com/thebiggame/bigbot/internal/helpers"
)

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
			{
				Name:        "infoboard",
				Description: "📽️ Transition to Infoboard (the default projector display).",
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
	content := "😶 Unknown command..."

	switch options[0].Name {
	case "status":
		if mod.goobsIsConnected() {
			content = "🙆 OBS is connected."
		} else {
			content = "🙅 OBS is **not connected.**"
		}
	case "ftb":
		if !mod.goobsIsConnected() {
			content = "🙅 OBS is **not connected.**"
			break
		}
		// Let the client know we're working on it.
		if helpers.DiscordDeferEphemeralInteraction(s, i) != nil {
			return true, err
		}
		return true, mod.discordCommandAVFTB(s, i)
	case "infoboard":
		if !mod.goobsIsConnected() {
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
}

func (mod *AVBridge) discordCommandAVFTB(s *discordgo.Session, i *discordgo.InteractionCreate) (err error) {
	// Set preview scene to black...
	_, err = mod.ws.Scenes.SetCurrentPreviewScene(&scenes.SetCurrentPreviewSceneParams{
		SceneName: &sceneBlack,
	})
	if err != nil {
		return err
	}

	// then transition to it.
	_, err = mod.ws.Transitions.SetCurrentSceneTransition(&transitions.SetCurrentSceneTransitionParams{
		TransitionName: &transitionFade,
	})
	if err != nil {
		return err
	}
	_, err = mod.ws.Transitions.TriggerStudioModeTransition(&transitions.TriggerStudioModeTransitionParams{})
	if err != nil {
		return err
	}

	// Finally, confirm we did the thing.
	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: "_Fading to black..._",
	})
	return err
}

func (mod *AVBridge) discordCommandAVInfoboard(s *discordgo.Session, i *discordgo.InteractionCreate) (err error) {
	// Set preview scene to Infoboard...
	_, err = mod.ws.Scenes.SetCurrentPreviewScene(&scenes.SetCurrentPreviewSceneParams{
		SceneName: &sceneDefault,
	})
	if err != nil {
		return err
	}

	// then transition to it.
	_, err = mod.ws.Transitions.SetCurrentSceneTransition(&transitions.SetCurrentSceneTransitionParams{
		TransitionName: &transitionStinger,
	})
	if err != nil {
		return err
	}
	_, err = mod.ws.Transitions.TriggerStudioModeTransition(&transitions.TriggerStudioModeTransitionParams{})
	if err != nil {
		return err
	}

	// Finally, confirm we did the thing.
	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: "Switched to Infoboard.",
	})
	return err
}

var defaultAVCommandPermissions int64 = discordgo.PermissionAdministrator
