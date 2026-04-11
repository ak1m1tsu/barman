package command

import "github.com/bwmarrin/discordgo"

func NewHelpCommand() (*discordgo.ApplicationCommand, Handler) {
	cmd := &discordgo.ApplicationCommand{
		Name:        "help",
		Description: "Список доступных команд",
	}

	handler := func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		embed := &discordgo.MessageEmbed{
			Title: "Доступные команды",
			Color: 0x5865F2,
			Fields: []*discordgo.MessageEmbedField{
				{Name: "/ping", Value: "Проверить задержку бота"},
				{Name: "/help", Value: "Показать этот список"},
				{Name: "/userinfo [пользователь]", Value: "Информация о пользователе"},
				{Name: "/autorole set <роль>", Value: "Установить авто-роль (требует ManageRoles)"},
				{Name: "/autorole remove", Value: "Удалить авто-роль (требует ManageRoles)"},
				{Name: "/autorole info", Value: "Показать текущую авто-роль (требует ManageRoles)"},
			},
		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{ //nolint:errcheck
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		})
	}

	return cmd, handler
}
