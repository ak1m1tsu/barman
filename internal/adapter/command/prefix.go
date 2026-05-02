package command

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	"github.com/ak1m1tsu/barman/internal/pkg/discordutil"
	guilduc "github.com/ak1m1tsu/barman/internal/usecase/guild"
)

// NewPrefixCommand returns the /prefix slash command and its handler.
// The handler shows the current guild prefix and presents buttons to change or reset it;
// requires the ManageGuild permission.
func NewPrefixCommand(getUC *guilduc.GetPrefixUseCase, timeout time.Duration) (*discordgo.ApplicationCommand, Handler) {
	cmd := &discordgo.ApplicationCommand{
		Name:        "prefix",
		Description: "Управление префиксом команд сервера",
	}

	handler := func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Member == nil {
			discordutil.RespondEphemeral(s, i, "Команда доступна только на сервере.")
			return
		}

		if i.Member.Permissions&discordgo.PermissionManageGuild == 0 {
			discordutil.RespondEphemeral(s, i, "Недостаточно прав. Требуется право **Управление сервером**.")
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		prefix, err := getUC.Execute(ctx, i.GuildID)
		if err != nil {
			logrus.WithError(err).WithField("guild_id", i.GuildID).Error("failed to get prefix")
			discordutil.RespondEphemeral(s, i, "Ошибка при получении префикса.")
			return
		}

		var currentDisplay string
		if prefix == "" {
			currentDisplay = "глобальный по умолчанию"
		} else {
			currentDisplay = fmt.Sprintf("`%s`", prefix)
		}

		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Текущий префикс: %s", currentDisplay),
				Flags:   discordgo.MessageFlagsEphemeral,
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{
								Label:    "Изменить",
								Style:    discordgo.PrimaryButton,
								CustomID: "prefix_set",
							},
							discordgo.Button{
								Label:    "Сбросить",
								Style:    discordgo.DangerButton,
								CustomID: "prefix_reset",
							},
						},
					},
				},
			},
		}); err != nil {
			logrus.WithError(err).WithField("guild_id", i.GuildID).Error("prefix: failed to send response")
		}
	}

	return cmd, handler
}
