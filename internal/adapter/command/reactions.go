package command

import (
	"context"
	"fmt"
	"slices"
	"sort"
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

		sfwKeys := make([]string, 0, len(reactionOrder))
		nsfwKeys := make([]string, 0)
		for _, key := range reactionOrder {
			if nsfwReactions[key] {
				nsfwKeys = append(nsfwKeys, key)
			} else {
				sfwKeys = append(sfwKeys, key)
			}
		}

		byCount := func(keys []string) {
			sort.SliceStable(keys, func(a, b int) bool {
				return stats[keys[a]] > stats[keys[b]]
			})
		}
		byCount(sfwKeys)
		byCount(nsfwKeys)

		writeSection := func(sb *strings.Builder, keys []string) {
			for _, key := range keys {
				meta := reactionsMeta[key]
				desc := strings.TrimPrefix(meta.withoutTarget, "%s ")
				fmt.Fprintf(sb, "`%-10s` — %s · **%d**\n", key, desc, stats[key])
			}
		}

		var sfwSb, nsfwSb strings.Builder
		writeSection(&sfwSb, sfwKeys)
		writeSection(&nsfwSb, nsfwKeys)

		var flags discordgo.MessageFlags
		if i.Member == nil || !slices.Contains(ownerIDs, i.Member.User.ID) {
			flags = discordgo.MessageFlagsEphemeral
		}

		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title: "Доступные реакции",
						Color: 0x5865F2,
						Fields: []*discordgo.MessageEmbedField{
							{Name: "SFW", Value: sfwSb.String(), Inline: false},
							{Name: "NSFW 🔞", Value: nsfwSb.String(), Inline: false},
						},
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
