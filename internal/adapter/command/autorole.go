package command

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	"github.com/ak1m1tsu/barman/internal/adapter/discordutil"
	guilduc "github.com/ak1m1tsu/barman/internal/usecase/guild"
)

// NewAutoRoleCommand returns the /autorole slash command and its handler.
// The handler shows the current auto-role and presents buttons to change or remove it;
// requires the ManageRoles permission.
func NewAutoRoleCommand(getUC *guilduc.GetAutoRoleUseCase, timeout time.Duration) (*discordgo.ApplicationCommand, Handler) {
	cmd := &discordgo.ApplicationCommand{
		Name:        "autorole",
		Description: "Управление авто-ролью сервера",
	}

	handler := func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Member == nil {
			discordutil.RespondEphemeral(s, i, "Команда доступна только на сервере.")
			return
		}

		if i.Member.Permissions&discordgo.PermissionManageRoles == 0 {
			discordutil.RespondEphemeral(s, i, "Недостаточно прав. Требуется право **Управление ролями**.")
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		g, err := getUC.Execute(ctx, i.GuildID)
		if err != nil {
			logrus.WithError(err).WithField("guild_id", i.GuildID).Error("failed to get autorole")
			discordutil.RespondEphemeral(s, i, "Ошибка при получении авто-роли.")
			return
		}

		var current string
		if g == nil || g.AutoRoleID == "" {
			current = "не установлена"
		} else {
			current = fmt.Sprintf("<@&%s>", g.AutoRoleID)
		}

		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content:    fmt.Sprintf("Текущая авто-роль: %s", current),
				Flags:      discordgo.MessageFlagsEphemeral,
				Components: []discordgo.MessageComponent{AutoRoleButtonsRow()},
			},
		}); err != nil {
			logrus.WithError(err).WithField("guild_id", i.GuildID).Error("autorole: failed to send response")
		}
	}

	return cmd, handler
}
