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
						MaxLength:   40,
					},
				},
			},
			{
				Name:        "alert-end",
				Description: "🔕 End the Alert early.",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
			{
				Name:        "announcement",
				Description: "🔔 Send an Announcement to all attendees. (This makes noise!)",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "name",
						Description: "A description of what's happening.",
						Required:    true,
						MaxLength:   200,
					},
				},
			},
			{
				Name:        "announcement-end",
				Description: "🔕 End the Announcement (return to normal service).",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
			{
				Name:        "schedule",
				Description: "📆 Update the Now & Next display.",
				Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Name:        "now",
						Description: "Set the CURRENT event.",
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:        discordgo.ApplicationCommandOptionString,
								Name:        "name",
								Description: "A brief description of the event.",
								Required:    true,
								MaxLength:   100,
							},
						},
					},
					{
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Name:        "next",
						Description: "Set the NEXT event.",
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:        discordgo.ApplicationCommandOptionString,
								Name:        "name",
								Description: "A brief description of the event. Leave blank to show no upcoming event.",
								Required:    false,
								MaxLength:   100,
							},
						},
					},
				},
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
		// Let the client know we're working on it.
		if helpers.DiscordDeferEphemeralInteraction(s, i) != nil {
			return true, err
		}
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
		_, err = helpers.DiscordInteractionFollowupMessage(s, i, "Alert Fired. Go be an attention whore!")
		return true, err
	case "alert-end":
		err := nodecg.MessageSend(*mod.ctx, "alert-end", nil)
		if err != nil {
			return true, err
		}
		_, err = helpers.DiscordInteractionFollowupMessage(s, i, "Alert Revoked.")
		return true, err
	case "announcement":
		// Let the client know we're working on it.
		if helpers.DiscordDeferEphemeralInteraction(s, i) != nil {
			return true, err
		}
		optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
		for _, opt := range options[0].Options {
			optionMap[opt.Name] = opt
		}
		name := "This is a service update from tBG Crew."

		if optionMap["name"] != nil {
			name = optionMap["name"].StringValue()
		}

		// First set the information body.
		err := nodecg.ReplicantSet(*mod.ctx, "event:info:body", name)
		if err != nil {
			return true, err
		}

		// Then set it to active (plays the announcement chime & displays it)
		err = nodecg.ReplicantSet(*mod.ctx, "event:info:active", true)
		if err != nil {
			return true, err
		}

		_, err = helpers.DiscordInteractionFollowupMessage(s, i, "ℹ Information is being displayed.")
		return true, err
	case "announcement-end":
		// Let the client know we're working on it.
		if helpers.DiscordDeferEphemeralInteraction(s, i) != nil {
			return true, err
		}
		err := nodecg.ReplicantSet(*mod.ctx, "event:info:active", false)
		if err != nil {
			return true, err
		}
		_, err = helpers.DiscordInteractionFollowupMessage(s, i, "ℹ Information update removed.")
		return true, err
	case "schedule":
		if helpers.DiscordDeferEphemeralInteraction(s, i) != nil {
			return true, err
		}
		switch options[0].Options[0].Name {
		case "now":
			// Set the "now" display.
			err := nodecg.ReplicantSet(*mod.ctx, "schedule:now", options[0].Options[0].Options[0].StringValue())
			if err != nil {
				return true, err
			}
			_, err = helpers.DiscordInteractionFollowupMessage(s, i, "👍 Schedule updated.")
			return true, err
		case "next":
			// Let the client know we're working on it.
			if helpers.DiscordDeferEphemeralInteraction(s, i) != nil {
				return true, err
			}
			// Set the "next" display.
			// Need to do a bounds check against the options.
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options[0].Options[0].Options {
				optionMap[opt.Name] = opt
			}
			var newEventValue string
			if optionMap["name"] != nil {
				newEventValue = optionMap["name"].StringValue()
			}
			err := nodecg.ReplicantSet(*mod.ctx, "schedule:next", newEventValue)
			if err != nil {
				return true, err
			}
			_, err = helpers.DiscordInteractionFollowupMessage(s, i, "👍 Schedule updated.")
			return true, err
		}
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
