package notifications

import (
	"github.com/bwmarrin/discordgo"
	"github.com/thebiggame/bigbot/internal/avbridge/ngtbg"
	"github.com/thebiggame/bigbot/internal/avcomms"
	"github.com/thebiggame/bigbot/internal/config"
	"github.com/thebiggame/bigbot/internal/helpers"
	"github.com/thebiggame/bigbot/internal/log"
	"strings"
)

func (mod *Notifications) HandleDiscordCommand(s *discordgo.Session, i *discordgo.InteractionCreate) (handled bool, err error) {
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
			err := avcomms.NodeCG.MessageSend(*mod.ctx, config.RuntimeConfig.AV.NodeCG.BundleName, ngtbg.NodeCGMessageChannelAlert, ngtbg.NodeCGMessageAlert{Name: name, Flair: flair})
			if err != nil {
				return true, err
			}
			_, err = helpers.DiscordInteractionFollowupMessage(s, i, "Alert Fired. Go be an attention whore!")
			return true, err
		case "alert-end":
			err := avcomms.NodeCG.MessageSend(*mod.ctx, config.RuntimeConfig.AV.NodeCG.BundleName, ngtbg.NodeCGMessageChannelAlertEnd, nil)
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
			err := avcomms.NodeCG.ReplicantSet(*mod.ctx, config.RuntimeConfig.AV.NodeCG.BundleName, ngtbg.NodeCGReplicantEventInfoActive, false)
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
			err := avcomms.NodeCG.ReplicantSet(*mod.ctx, config.RuntimeConfig.AV.NodeCG.BundleName, ngtbg.NodeCGReplicantEventInfoBody, name)
			if err != nil {
				// NodeCG not available for some reason.
				log.Infof("NodeCG not available: %s", err)
			} else {
				// Then set it to active (plays the announcement chime & displays it)
				err = avcomms.NodeCG.ReplicantSet(*mod.ctx, config.RuntimeConfig.AV.NodeCG.BundleName, ngtbg.NodeCGReplicantEventInfoActive, true)
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
