package avbridge

import (
	"github.com/andreykaipov/goobs/api/requests/scenes"
	"github.com/andreykaipov/goobs/api/requests/transitions"
	"github.com/bwmarrin/discordgo"
	"github.com/thebiggame/bigbot/internal/avbridge/ngtbg"
	"github.com/thebiggame/bigbot/internal/config"
	"github.com/thebiggame/bigbot/internal/helpers"
	"strings"
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
				Description: "🔔 Open the Announcement modal to create a new Announcement. (Submitting this makes noise!)",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
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
			err := mod.ncg.MessageSend(*mod.ctx, config.RuntimeConfig.AV.NodeCG.BundleName, ngtbg.NodeCGMessageChannelAlert, ngtbg.NodeCGMessageAlert{Name: name, Flair: flair})
			if err != nil {
				return true, err
			}
			_, err = helpers.DiscordInteractionFollowupMessage(s, i, "Alert Fired. Go be an attention whore!")
			return true, err
		case "alert-end":
			err := mod.ncg.MessageSend(*mod.ctx, config.RuntimeConfig.AV.NodeCG.BundleName, ngtbg.NodeCGMessageChannelAlertEnd, nil)
			if err != nil {
				return true, err
			}
			_, err = helpers.DiscordInteractionFollowupMessage(s, i, "Alert Revoked.")
			return true, err
		case "announcement":
			// Pop a modal to continue the interaction.
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseModal,
				Data: &discordgo.InteractionResponseData{
					CustomID: "bigbot_avbridge_announcement_" + i.Interaction.Member.User.ID,
					Title:    "Update Announcement",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:  "body",
									Label:     "Your announcement",
									Style:     discordgo.TextInputParagraph,
									Required:  true,
									MinLength: 1,
									MaxLength: 250,
								},
							},
						},
					},
				},
			})
			return true, err
		case "announcement-end":
			// Let the client know we're working on it.
			if helpers.DiscordDeferEphemeralInteraction(s, i) != nil {
				return true, err
			}
			err := mod.ncg.ReplicantSet(*mod.ctx, config.RuntimeConfig.AV.NodeCG.BundleName, ngtbg.NodeCGReplicantEventInfoActive, false)
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
				err := mod.ncg.ReplicantSet(*mod.ctx, config.RuntimeConfig.AV.NodeCG.BundleName, ngtbg.NodeCGReplicantScheduleNow, options[0].Options[0].Options[0].StringValue())
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
				err := mod.ncg.ReplicantSet(*mod.ctx, config.RuntimeConfig.AV.NodeCG.BundleName, ngtbg.NodeCGReplicantScheduleNext, newEventValue)
				if err != nil {
					return true, err
				}
				_, err = helpers.DiscordInteractionFollowupMessage(s, i, "👍 Schedule updated.")
				return true, err
			}
		}

		// Not handled by specific handler function, respond with content data.
		return true, helpers.DiscordInteractionEphemeralResponse(s, i, content)
	case discordgo.InteractionModalSubmit:
		// Modal submission.
		data := i.ModalSubmitData()

		switch {
		case strings.HasPrefix(data.CustomID, "bigbot_avbridge_announcement_"):
			// Data has returned from the Announcement modal.
			if helpers.DiscordDeferEphemeralInteraction(s, i) != nil {
				return true, err
			}

			// Potentially unsafe? This is how the example does it.
			name := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

			// First set the information body.
			err := mod.ncg.ReplicantSet(*mod.ctx, config.RuntimeConfig.AV.NodeCG.BundleName, ngtbg.NodeCGReplicantEventInfoBody, name)
			if err != nil {
				return true, err
			}

			// Then set it to active (plays the announcement chime & displays it)
			err = mod.ncg.ReplicantSet(*mod.ctx, config.RuntimeConfig.AV.NodeCG.BundleName, ngtbg.NodeCGReplicantEventInfoActive, true)
			if err != nil {
				return true, err
			}

			_, err = helpers.DiscordInteractionFollowupMessage(s, i, "ℹ Information is being displayed.")
			return true, err
		default:
			// This isn't anything to do with us.
			return false, nil
		}
		return true, err
	default:
		// Not something we recognise.
		return false, nil
	}

}

func (mod *AVBridge) discordCommandAVFTB(s *discordgo.Session, i *discordgo.InteractionCreate) (err error) {
	// Set preview scene to black...
	_, err = mod.ws.Scenes.SetCurrentPreviewScene(&scenes.SetCurrentPreviewSceneParams{
		SceneName: &ngtbg.OBSSceneBlack,
	})
	if err != nil {
		return err
	}

	// then transition to it.
	_, err = mod.ws.Transitions.SetCurrentSceneTransition(&transitions.SetCurrentSceneTransitionParams{
		TransitionName: &ngtbg.OBSTransFade,
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
		SceneName: &ngtbg.OBSSceneDefault,
	})
	if err != nil {
		return err
	}

	// then transition to it.
	_, err = mod.ws.Transitions.SetCurrentSceneTransition(&transitions.SetCurrentSceneTransitionParams{
		TransitionName: &ngtbg.OBSTransStingModernWipe,
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
