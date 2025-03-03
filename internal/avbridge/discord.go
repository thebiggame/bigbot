package avbridge

import (
	"github.com/andreykaipov/goobs/api/requests/scenes"
	"github.com/andreykaipov/goobs/api/requests/transitions"
	"github.com/bwmarrin/discordgo"
	"github.com/thebiggame/bigbot/internal/helpers"
	"github.com/thebiggame/bigbot/pkg/nodecg"
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
			{
				Name:        "alert",
				Description: "🔔 Sound an Alert on the AV system.",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionBoolean,
						Name:        "flair",
						Description: "Whether the alert should arrive with 'flair'. WARNING - this makes noise!",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "name",
						Description: "A short description of why you want people's attention.",
						Required:    true,
						MaxLength:   50,
					},
				},
			},
			{
				Name:        "alert-end",
				Description: "🔕 End the Alert early.",
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
	case "alert":
		optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
		for _, opt := range options[0].Options {
			optionMap[opt.Name] = opt
		}
		name := "Pay Attention!"
		var flair bool

		if optionMap["name"] != nil {
			name = optionMap["name"].StringValue()
		}
		if optionMap["flair"] != nil {
			flair = optionMap["flair"].BoolValue()
		}
		err := nodecg.MessageSend(*mod.ctx, "alert", ngtbgNodeCGMessageAlert{Name: name, Flair: flair})
		if err != nil {
			return true, err
		}
		return true, helpers.DiscordInteractionEphemeralResponse(s, i, "Alert Fired.")
	case "alert-end":
		err := nodecg.MessageSend(*mod.ctx, "alert-end", nil)
		if err != nil {
			return true, err
		}
		return true, helpers.DiscordInteractionEphemeralResponse(s, i, "Alert Rescinded.")
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
	_, err = helpers.DiscordInteractionFollowupMessage(s, i, "_Fading to black_...")
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
	_, err = helpers.DiscordInteractionFollowupMessage(s, i, "Switched to Infoboard.")
	return err
}

var defaultAVCommandPermissions int64 = discordgo.PermissionAdministrator
var defaultAVCommandDMPermissions = false
