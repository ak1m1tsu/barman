package command

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func NewPingCommand() (*discordgo.ApplicationCommand, Handler) {
	cmd := &discordgo.ApplicationCommand{
		Name:        "ping",
		Description: "Проверить задержку бота",
	}

	handler := func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		latency := s.HeartbeatLatency().Round(time.Millisecond)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{ //nolint:errcheck
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Pong! Задержка: %s", latency),
			},
		})
	}

	return cmd, handler
}
