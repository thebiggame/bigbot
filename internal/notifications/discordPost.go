package notifications

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/thebiggame/bigbot/internal/config"
	"github.com/thebiggame/bigbot/internal/log"
	"strings"
)

// This file handles sending messages to Discord channels as necessary.

func sendNotificationToDiscord(s *discordgo.Session, message string) (err error) {
	// Get Channel ID from config
	var channelID = config.RuntimeConfig.Discord.Announcements.ChannelID
	if config.RuntimeConfig.Discord.Announcements.ChannelID == "" {
		log.Info("No Notification Channel ID set, not sending notification to Discord")
		return
	}

	// Build the message.
	var msg = ":information_source: **Message from tBG Crew: **\n" +
		"> %s"
	// Prepend line breaks in message with quotation markdown.
	message = strings.ReplaceAll(message, "\n", "\n> ")

	_, err = s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: fmt.Sprintf(msg, message),
		TTS:     true,
	})
	return err
}
