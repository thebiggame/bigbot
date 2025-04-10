package musicparty

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/thebiggame/bigbot/internal/avbridge/ngtbg"
	"github.com/thebiggame/bigbot/internal/avcomms"
	"github.com/thebiggame/bigbot/internal/config"
	"github.com/thebiggame/bigbot/internal/helpers"
)

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "music",
		Description: "🎵 Find out what's playing.",
	},
}

func (mod *MusicParty) DiscordHandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) (handled bool, err error) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		// Handle normally.
		if i.ApplicationCommandData().Name != "music" {
			return false, nil
		}

		// Get music data (if available)
		// Right now this just calls out to NodeCG. Going forward, we'll do this with MusicParty directly.
		var data ngtbg.NodeCGReplicantDataMusicData
		err = avcomms.NodeCG.ReplicantGetDecode(*mod.ctx, config.RuntimeConfig.AV.NodeCG.BundleName, ngtbg.NodeCGReplicantMusicData, &data)
		if err != nil {
			return true, err
		}
		musicString := fmt.Sprintf("🎵 **%s** / %s", data.Title, data.Artist)
		return true, helpers.DiscordInteractionEphemeralResponse(s, i, musicString)
	default:
		return false, nil
	}
}
