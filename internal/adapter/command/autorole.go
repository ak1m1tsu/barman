package command

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	guilduc "github.com/ak1m1tsu/barman/internal/usecase/guild"
)

func NewAutoRoleCommand(getUC *guilduc.GetAutoRoleUseCase) (*discordgo.ApplicationCommand, Handler) {
	cmd := &discordgo.ApplicationCommand{
		Name:        "autorole",
		Description: "Управление авто-ролью сервера",
	}

	handler := func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Member == nil {
			respondEphemeral(s, i, "Команда доступна только на сервере.")
			return
		}

		if i.Member.Permissions&discordgo.PermissionManageRoles == 0 {
			respondEphemeral(s, i, "Недостаточно прав. Требуется право **Управление ролями**.")
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		g, err := getUC.Execute(ctx, i.GuildID)
		if err != nil {
			logrus.WithError(err).WithField("guild_id", i.GuildID).Error("failed to get autorole")
			respondEphemeral(s, i, "Ошибка при получении авто-роли.")
			return
		}

		var current string
		if g == nil || g.AutoRoleID == "" {
			current = "не установлена"
		} else {
			current = fmt.Sprintf("<@&%s>", g.AutoRoleID)
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{ //nolint:errcheck
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Текущая авто-роль: %s", current),
				Flags:   discordgo.MessageFlagsEphemeral,
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{
								Label:    "Изменить",
								Style:    discordgo.PrimaryButton,
								CustomID: "autorole_set",
							},
							discordgo.Button{
								Label:    "Удалить",
								Style:    discordgo.DangerButton,
								CustomID: "autorole_remove",
							},
							discordgo.Button{
								Label:    "Отменить",
								Style:    discordgo.SecondaryButton,
								CustomID: "autorole_cancel",
							},
						},
					},
				},
			},
		})
	}

	return cmd, handler
}

func respondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{ //nolint:errcheck
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
