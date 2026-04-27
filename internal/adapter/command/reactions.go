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

// NewReactionsCommand returns the /reactions slash command and its handler.
// The handler shows a paginated embed with one reaction card per page.
// The response is public for owner IDs and ephemeral for all other users.
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

		userID := ""
		if i.Member != nil {
			userID = i.Member.User.ID
		}

		var flags discordgo.MessageFlags
		if i.Member == nil || !slices.Contains(ownerIDs, i.Member.User.ID) {
			flags = discordgo.MessageFlagsEphemeral
		}

		embed := BuildReactionCard(0, stats)
		buttons := BuildCardButtons(0, userID, len(ReactionOrder))

		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{embed},
				Components: []discordgo.MessageComponent{buttons},
				Flags:      flags,
			},
		}); err != nil {
			log.WithError(err).Error("reactions: failed to send response")
		}
	}

	return cmd, handler
}

// BuildReactionCard builds an embed for a single reaction card at the given index.
func BuildReactionCard(index int, stats map[string]int64) *discordgo.MessageEmbed {
	key := ReactionOrder[index]
	meta := ReactionsMeta[key]
	emoji := ReactionEmojis[key]

	desc := strings.TrimPrefix(meta.WithoutTarget, "%s ")
	example := fmt.Sprintf(meta.WithTarget, "Алиса", "Боба")
	count := stats[key]

	return &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s %s", emoji, key),
		Color: ColorDiscordBranding,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Команда", Value: fmt.Sprintf("`/react %s`  или  `!%s`", key, key), Inline: false},
			{Name: "Описание", Value: desc, Inline: true},
			{Name: "Использований", Value: fmt.Sprintf("%d", count), Inline: true},
			{Name: "Пример", Value: example, Inline: false},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Карточка %d из %d", index+1, len(ReactionOrder)),
		},
	}
}

// BuildCardButtons returns the action row for card navigation.
func BuildCardButtons(index int, userID string, total int) discordgo.ActionsRow {
	prevDisabled := index == 0
	nextDisabled := index == total-1

	prevStyle := discordgo.PrimaryButton
	if prevDisabled {
		prevStyle = discordgo.SecondaryButton
	}
	nextStyle := discordgo.PrimaryButton
	if nextDisabled {
		nextStyle = discordgo.SecondaryButton
	}

	return discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "←",
				Style:    prevStyle,
				Disabled: prevDisabled,
				CustomID: fmt.Sprintf("reactions:prev:%s:%d", userID, index),
			},
			discordgo.Button{
				Label:    "Весь список",
				Style:    discordgo.SecondaryButton,
				CustomID: fmt.Sprintf("reactions:all:%s:%d", userID, index),
			},
			discordgo.Button{
				Label:    "→",
				Style:    nextStyle,
				Disabled: nextDisabled,
				CustomID: fmt.Sprintf("reactions:next:%s:%d", userID, index),
			},
		},
	}
}

// BuildTableEmbed builds an embed listing all reactions sorted by usage count.
func BuildTableEmbed(userID string, fromIndex int, stats map[string]int64) (*discordgo.MessageEmbed, discordgo.ActionsRow) {
	type row struct {
		key   string
		count int64
	}
	rows := make([]row, len(ReactionOrder))
	for i, key := range ReactionOrder {
		rows[i] = row{key, stats[key]}
	}
	// Sort by count descending, stable to keep ReactionOrder as tiebreaker.
	for i := 1; i < len(rows); i++ {
		for j := i; j > 0 && rows[j].count > rows[j-1].count; j-- {
			rows[j], rows[j-1] = rows[j-1], rows[j]
		}
	}

	const nameW, descW = 10, 21
	var sb strings.Builder
	fmt.Fprintf(&sb, "%-*s  %-*s  %s\n", nameW, "Реакция", descW, "Действие", "Раз")
	sb.WriteString(strings.Repeat("─", nameW+descW+9) + "\n")
	for _, r := range rows {
		meta := ReactionsMeta[r.key]
		desc := strings.TrimPrefix(meta.WithoutTarget, "%s ")
		fmt.Fprintf(&sb, "%-*s  %-*s  %d\n", nameW, r.key, descW, desc, r.count)
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Все реакции",
		Description: "```\n" + sb.String() + "```",
		Color:       0x5865F2,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Используй /react <тип> или !<тип>",
		},
	}

	buttons := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Вернуться",
				Style:    discordgo.SecondaryButton,
				CustomID: fmt.Sprintf("reactions:back:%s:%d", userID, fromIndex),
			},
		},
	}

	return embed, buttons
}

// ReactionEmojis maps reaction type → emoji for card display.
var ReactionEmojis = map[string]string{
	"hug":       "🤗",
	"pat":       "🐾",
	"kiss":      "💋",
	"cuddle":    "🫂",
	"feed":      "🍴",
	"wave":      "👋",
	"wink":      "😉",
	"smile":     "😊",
	"highfive":  "🙌",
	"handshake": "🤝",
	"poke":      "👉",
	"tickle":    "🤭",
	"lick":      "👅",
	"bite":      "😬",
	"slap":      "🖐️",
	"punch":     "👊",
	"love":      "❤️",
	"nuzzle":    "😽",
	"shy":       "😳",
	"nervous":   "😰",
	"nosebleed": "🩸",
	"brofist":   "🤜",
	"headbang":  "🤘",
	"sad":       "😢",
	"peek":      "👀",
}
