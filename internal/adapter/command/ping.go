package command

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

func NewPingCommand() (*discordgo.ApplicationCommand, Handler) {
	cmd := &discordgo.ApplicationCommand{
		Name:        "ping",
		Description: "Проверить задержку бота",
	}

	handler := func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		latency := s.HeartbeatLatency().Round(time.Millisecond)
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Pong! Задержка: %s", latency),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}); err != nil {
			logrus.WithError(err).WithField("guild_id", i.GuildID).Error("ping: failed to send response")
		}
	}

	return cmd, handler
}
