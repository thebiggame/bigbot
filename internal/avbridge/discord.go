package avbridge

import (
	"github.com/andreykaipov/goobs/api/requests/scenes"
	"github.com/andreykaipov/goobs/api/requests/transitions"
	"github.com/bwmarrin/discordgo"
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
	content := ""

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
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			return true, err
		}

		// Set preview scene to black...
		_, err = mod.ws.Scenes.SetCurrentPreviewScene(&scenes.SetCurrentPreviewSceneParams{
			SceneName: &sceneBlack,
		})
		if err != nil {
			return true, err
		}

		// then transition to it.
		_, err = mod.ws.Transitions.SetCurrentSceneTransition(&transitions.SetCurrentSceneTransitionParams{
			TransitionName: &transitionFade,
		})
		if err != nil {
			return true, err
		}
		_, err = mod.ws.Transitions.TriggerStudioModeTransition(&transitions.TriggerStudioModeTransitionParams{})
		if err != nil {
			return true, err
		}

		// Finally, confirm we did the thing.
		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "_Fading to black..._",
		})
		if err != nil {
			return true, err
		}
		return true, nil
	case "infoboard":
		if !mod.goobsIsConnected() {
			content = "🙅 OBS is **not connected.**"
			break
		}
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			return true, err
		}

		// Set preview scene to Infoboard...
		_, err = mod.ws.Scenes.SetCurrentPreviewScene(&scenes.SetCurrentPreviewSceneParams{
			SceneName: &sceneDefault,
		})
		if err != nil {
			return true, err
		}

		// then transition to it.
		_, err = mod.ws.Transitions.SetCurrentSceneTransition(&transitions.SetCurrentSceneTransitionParams{
			TransitionName: &transitionStinger,
		})
		if err != nil {
			return true, err
		}
		_, err = mod.ws.Transitions.TriggerStudioModeTransition(&transitions.TriggerStudioModeTransitionParams{})
		if err != nil {
			return true, err
		}

		// Finally, confirm we did the thing.
		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Switched to Infoboard.",
		})
		if err != nil {
			return true, err
		}
		return true, nil
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
