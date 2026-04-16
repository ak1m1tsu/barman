package command

import (
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

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
				{Name: "/autorole", Value: "Управление авто-ролью сервера (требует ManageRoles)"},
				{Name: "/react <тип> [пользователь]", Value: "Отправить аниме-реакцию"},
				{Name: "/reactions", Value: "Список доступных типов реакций"},
				{Name: "/prefix", Value: "Управление префиксом команд сервера (требует ManageServer)"},
				{Name: "!<тип> [@пользователь]", Value: "Отправить реакцию через префи��с; в reply — цель определяется автоматически"},
			},
		}
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		}); err != nil {
			logrus.WithError(err).WithField("guild_id", i.GuildID).Error("help: failed to send response")
		}
	}

	return cmd, handler
}
