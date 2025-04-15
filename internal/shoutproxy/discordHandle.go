package shoutproxy

import (
	"github.com/bwmarrin/discordgo"
	"github.com/thebiggame/bigbot/internal/avbridge/ngtbg"
	"github.com/thebiggame/bigbot/internal/avcomms"
	"github.com/thebiggame/bigbot/internal/config"
	"github.com/thebiggame/bigbot/internal/log"
	"time"
)

// Forward messages in the configured channel to NodeCG.
func (mod *ShoutProxy) DiscordHandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) (err error) {
	// Ensure the shout happened in the expected target guild and channel.
	var channelID = config.RuntimeConfig.Discord.Shoutbox.ChannelID
	if channelID == "" {
		log.Debug("No Shoutbox Channel ID set, not dispatching")
		return
	}

	if m.Message.ChannelID == channelID && m.Message.GuildID == config.RuntimeConfig.Discord.GuildID {
		// This message is a shout! Let's make it known to NodeCG.

		var userName = m.Message.Author.GlobalName
		if m.Message.Member.Nick != "" {
			// The user has a server-specific display name set. Use that.
			userName = m.Message.Member.Nick
		}
		var userAvatar = m.Message.Author.AvatarURL("128x128")
		if m.Message.Member.Avatar != "" {
			// The user has a server-specific avatar set. Use that.
			userAvatar = m.Message.Member.AvatarURL("128x128")
		}

		var shoutEntry = ngtbg.NodeCGReplicantDataShoutboxEntry{
			ID: "DISC-" + m.Message.ID,
			User: struct {
				Name   string `json:"name"`
				Avatar string `json:"avatar_url"`
			}{
				Name:   userName,
				Avatar: userAvatar,
			},
			Timestamp: m.Message.Timestamp.Format(time.RFC3339),
			Message:   m.Message.Content,
		}
		err := avcomms.NodeCG.MessageSend(*mod.ctx, config.RuntimeConfig.AV.NodeCG.BundleName, ngtbg.NodeCGMessageShoutboxNew, shoutEntry)
		// err = avcomms.NodeCG.ReplicantSet(*mod.ctx, config.RuntimeConfig.AV.NodeCG.BundleName, ngtbg.NodeCGReplicantShoutbox, shoutboxEntries)
		return err
	}

	return nil
}

func (mod *ShoutProxy) DiscordHandleInteraction(_ *discordgo.Session, _ *discordgo.InteractionCreate) (handled bool, err error) {
	return false, nil
}
