package command

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func NewReactionsCommand() (*discordgo.ApplicationCommand, Handler) {
	cmd := &discordgo.ApplicationCommand{
		Name:        "reactions",
		Description: "Список доступных типов реакций",
	}

	handler := func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		list := make([]string, 0, len(reactionOrder))
		for _, r := range reactionOrder {
			list = append(list, "`"+r+"`")
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{ //nolint:errcheck
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       "Доступные реакции",
						Description: strings.Join(list, " "),
						Color:       0x5865F2,
						Footer: &discordgo.MessageEmbedFooter{
							Text: "Используй /react <тип> или !<тип>",
						},
					},
				},
			},
		})
	}

	return cmd, handler
}
