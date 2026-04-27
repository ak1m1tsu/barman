package handler

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	"github.com/ak1m1tsu/barman/internal/adapter/command"
	reactionuc "github.com/ak1m1tsu/barman/internal/usecase/reaction"
)

// NewReactionsInteractionHandler handles button clicks for the /reactions paginated UI.
// CustomID format:
//
//	reactions:prev:{userID}:{index}   — go to previous card
//	reactions:next:{userID}:{index}   — go to next card
//	reactions:all:{userID}:{index}    — switch to table view (remembers index)
//	reactions:back:{userID}:{index}   — return from table to card at index
func NewReactionsInteractionHandler(getStats *reactionuc.GetStatsUseCase, timeout time.Duration) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionMessageComponent {
			return
		}

		customID := i.MessageComponentData().CustomID
		if !strings.HasPrefix(customID, "reactions:") {
			return
		}

		parts := strings.SplitN(customID, ":", 4)
		if len(parts) != 4 {
			return
		}
		action, userID, indexStr := parts[1], parts[2], parts[3]

		// Only the original invoker may interact.
		if i.Member == nil || i.Member.User.ID != userID {
			if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Это не ваша панель реакций.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			}); err != nil {
				return
			}
			return
		}

		index, err := strconv.Atoi(indexStr)
		if err != nil || index < 0 || index >= len(command.ReactionOrder) {
			return
		}

		log := logrus.WithFields(logrus.Fields{
			"guild_id": i.GuildID,
			"user_id":  userID,
			"action":   action,
			"index":    index,
			"command":  "reactions (interaction)",
		})

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		stats, err := getStats.Execute(ctx)
		if err != nil {
			log.WithError(err).Error("reactions interaction: failed to fetch stats")
			stats = map[string]int64{}
		}

		var embed *discordgo.MessageEmbed
		var row discordgo.ActionsRow

		switch action {
		case "prev":
			if index > 0 {
				index--
			}
			embed = command.BuildReactionCard(index, stats)
			row = command.BuildCardButtons(index, userID, len(command.ReactionOrder))
		case "next":
			if index < len(command.ReactionOrder)-1 {
				index++
			}
			embed = command.BuildReactionCard(index, stats)
			row = command.BuildCardButtons(index, userID, len(command.ReactionOrder))
		case "all":
			embed, row = command.BuildTableEmbed(userID, index, stats)
		case "back":
			embed = command.BuildReactionCard(index, stats)
			row = command.BuildCardButtons(index, userID, len(command.ReactionOrder))
		default:
			return
		}

		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{embed},
				Components: []discordgo.MessageComponent{row},
			},
		}); err != nil {
			log.WithError(err).Error("reactions interaction: failed to update message")
		}
	}
}
