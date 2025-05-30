package notifications

import (
	"github.com/bwmarrin/discordgo"
	"github.com/thebiggame/bigbot/internal/avbridge/ngtbg"
	bridge_wan "github.com/thebiggame/bigbot/internal/bridge-wan"
	"github.com/thebiggame/bigbot/internal/config"
	"github.com/thebiggame/bigbot/internal/helpers"
	"log/slog"
	"strings"
)

var commands = []*discordgo.ApplicationCommand{
	{
		Name:                     "notify",
		Description:              "🛎️ Send & manage notifications (you must be a crew member)",
		DefaultMemberPermissions: &defaultCrewCommandPermissions,
		DMPermission:             &defaultCrewCommandDMPermissions,
		Options: []*discordgo.ApplicationCommandOption{
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
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "delay",
						Description: "The amount of time to pause before displaying the alert. Defaults to Instant.",
						Required:    false,
						MaxValue:    15,
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
		},
	},
}

func (mod *Notifications) DiscordHandleMessage(session *discordgo.Session, message *discordgo.MessageCreate) (err error) {
	return nil
}

func (mod *Notifications) DiscordHandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) (handled bool, err error) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		// Handle normally.
		if i.ApplicationCommandData().Name != "notify" {
			return false, nil
		}
		options := i.ApplicationCommandData().Options
		content := "😶 Unknown command..."

		switch options[0].Name {
		case "alert":
			// Check the bridge is available.
			if !bridge_wan.BridgeIsAvailable() {
				return true, helpers.DiscordInteractionEphemeralResponse(s, i, "👻 **Event Bridge is not available**")
			}
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
			var delay uint64 = 0

			if optionMap["name"] != nil {
				name = optionMap["name"].StringValue()
			}
			if optionMap["flair"] != nil {
				flair = optionMap["flair"].BoolValue()
			}
			if optionMap["delay"] != nil {
				delay = optionMap["delay"].UintValue()
			}
			err := bridge_wan.EventBridge.BrReplicantSet(config.RuntimeConfig.AV.NodeCG.BundleName, ngtbg.NodeCGReplicantNotificationAlertData, ngtbg.NodeCGReplicantDataAlertData{
				Body:  name,
				Flair: flair,
				Delay: int(delay),
			})
			if err != nil {
				return true, err
			}
			err = bridge_wan.EventBridge.BrReplicantSet(config.RuntimeConfig.AV.NodeCG.BundleName, ngtbg.NodeCGReplicantNotificationAlertActive, true)
			if err != nil {
				return true, err
			}
			_, err = helpers.DiscordInteractionFollowupMessage(s, i, "Alert Fired. Go be an attention whore!")
			return true, err
		case "alert-end":
			// Let the client know we're working on it.
			if helpers.DiscordDeferEphemeralInteraction(s, i) != nil {
				return true, err
			}
			err = bridge_wan.EventBridge.BrReplicantSet(config.RuntimeConfig.AV.NodeCG.BundleName, ngtbg.NodeCGReplicantNotificationAlertActive, false)
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
					CustomID: "bigbot_notify_announcement_" + i.Interaction.Member.User.ID,
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
			err := bridge_wan.EventBridge.BrReplicantSet(config.RuntimeConfig.AV.NodeCG.BundleName, ngtbg.NodeCGReplicantEventInfoActive, false)
			if err != nil {
				return true, err
			}
			_, err = helpers.DiscordInteractionFollowupMessage(s, i, "ℹ Information update removed.")
			return true, err
		}

		// Not handled by specific handler function, respond with content data.
		return true, helpers.DiscordInteractionEphemeralResponse(s, i, content)
	case discordgo.InteractionModalSubmit:
		// Modal submission.
		data := i.ModalSubmitData()

		switch {
		case strings.HasPrefix(data.CustomID, "bigbot_notify_announcement_"):
			// Data has returned from the Announcement modal.
			if helpers.DiscordDeferEphemeralInteraction(s, i) != nil {
				return true, err
			}

			// Potentially unsafe? This is how the example does it.
			name := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

			// First attempt to set the information body.
			err := bridge_wan.EventBridge.BrReplicantSet(config.RuntimeConfig.AV.NodeCG.BundleName, ngtbg.NodeCGReplicantEventInfoBody, name)
			if err != nil {
				// NodeCG not available for some reason.
				logger.Info("NodeCG not available", slog.Any("error", err))
			} else {
				// Then set it to active (plays the announcement chime & displays it)
				err = bridge_wan.EventBridge.BrReplicantSet(config.RuntimeConfig.AV.NodeCG.BundleName, ngtbg.NodeCGReplicantEventInfoActive, true)
				if err != nil {
					return true, err
				}
			}

			// Separately, regardless of whether NodeCG is available or not, send to Discord channel (if configured).
			err = sendNotificationToDiscord(s, name)
			if err != nil {
				return true, err
			}

			_, err = helpers.DiscordInteractionFollowupMessage(s, i, "ℹ Information sent successfully.")
			return true, err
		default:
			// This isn't anything to do with us.
			return false, nil
		}
	default:
		// Not something we recognise.
		return false, nil
	}

}

var defaultCrewCommandPermissions int64 = discordgo.PermissionAdministrator
var defaultCrewCommandDMPermissions = false
