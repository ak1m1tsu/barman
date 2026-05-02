package discordutil

import (
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

// RespondEphemeral sends an ephemeral channel-message response to an interaction.
// Errors are logged but not returned; callers should not retry on failure.
func RespondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}); err != nil {
		logrus.WithError(err).WithField("guild_id", i.GuildID).Error("failed to send ephemeral response")
	}
}
