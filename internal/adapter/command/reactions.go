package command

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	reactionuc "github.com/ak1m1tsu/barman/internal/usecase/reaction"
)

func NewReactionsCommand(getStats *reactionuc.GetStatsUseCase, ownerIDs []string, timeout time.Duration) (*discordgo.ApplicationCommand, Handler) {
	cmd := &discordgo.ApplicationCommand{
		Name:        "reactions",
		Description: "Список доступных реакций с описанием и статистикой использования",
	}

	handler := func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		log := logrus.WithFields(logrus.Fields{
			"guild_id": i.GuildID,
			"command":  "reactions",
		})

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		stats, err := getStats.Execute(ctx)
		if err != nil {
			log.WithError(err).Error("reactions: failed to fetch stats")
			stats = map[string]int64{}
		}

		var sb strings.Builder
		for _, key := range reactionOrder {
			meta := reactionsMeta[key]
			desc := strings.TrimPrefix(meta.withoutTarget, "%s ")
			count := stats[key]
			fmt.Fprintf(&sb, "`%-10s` — %s · **%d**\n", key, desc, count)
		}

		var flags discordgo.MessageFlags
		if i.Member == nil || !slices.Contains(ownerIDs, i.Member.User.ID) {
			flags = discordgo.MessageFlagsEphemeral
		}

		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       "Доступные реакции",
						Description: sb.String(),
						Color:       0x5865F2,
						Footer: &discordgo.MessageEmbedFooter{
							Text: "Используй /react <тип> или !<тип>",
						},
					},
				},
				Flags: flags,
			},
		}); err != nil {
			log.WithError(err).Error("reactions: failed to send response")
		}
	}

	return cmd, handler
}
