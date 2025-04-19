package bridge_wan

import (
	"github.com/bwmarrin/discordgo"
)

func (mod *BridgeWAN) DiscordCommands() ([]*discordgo.ApplicationCommand, error) {
	return nil, nil
}

func (mod *BridgeWAN) DiscordHandleMessage(session *discordgo.Session, message *discordgo.MessageCreate) (err error) {
	return nil
}

func (mod *BridgeWAN) DiscordHandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) (handled bool, err error) {
	return false, nil
}
