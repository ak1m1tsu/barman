package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	"github.com/ak1m1tsu/barman/internal/adapter/command"
	"github.com/ak1m1tsu/barman/internal/pkg/discordutil"
	guilduc "github.com/ak1m1tsu/barman/internal/usecase/guild"
)

// NewAutoRoleInteractionHandler handles button clicks and role selection for /autorole.
func NewAutoRoleInteractionHandler(
	setUC *guilduc.SetAutoRoleUseCase,
	getUC *guilduc.GetAutoRoleUseCase,
	removeUC *guilduc.RemoveAutoRoleUseCase,
	timeout time.Duration,
) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionMessageComponent {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		data := i.MessageComponentData()
		log := logrus.WithFields(logrus.Fields{
			"guild_id":  i.GuildID,
			"custom_id": data.CustomID,
			"command":   "autorole (interactive)",
		})
		if i.Member != nil && i.Member.User != nil {
			log = log.WithField("user_id", i.Member.User.ID)
		}

		switch data.CustomID {
		case command.AutoRoleSetButtonID:
			// Swap buttons for a Role SelectMenu in-place
			if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseUpdateMessage,
				Data: &discordgo.InteractionResponseData{
					Content: "Выберите роль для авто-выдачи новым участникам:",
					Flags:   discordgo.MessageFlagsEphemeral,
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.SelectMenu{
									MenuType:    discordgo.RoleSelectMenu,
									CustomID:    command.AutoRoleSelectID,
									Placeholder: "Выберите роль...",
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									Label:    "Отменить",
									Style:    discordgo.SecondaryButton,
									CustomID: command.AutoRoleCancelButtonID,
								},
							},
						},
					},
				},
			}); err != nil {
				log.WithError(err).Error("autorole: failed to show role select")
			}

		case command.AutoRoleCancelButtonID:
			if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseUpdateMessage,
				Data: &discordgo.InteractionResponseData{
					Content:    "Отменено.",
					Flags:      discordgo.MessageFlagsEphemeral,
					Components: []discordgo.MessageComponent{},
				},
			}); err != nil {
				log.WithError(err).Error("autorole: failed to send cancel response")
			}

		case command.AutoRoleRemoveButtonID:
			if err := removeUC.Execute(ctx, i.GuildID); err != nil {
				log.WithError(err).Error("failed to remove autorole")
				discordutil.RespondEphemeral(s, i, "Ошибка при удалении авто-роли.")
				return
			}
			log.WithField("notify", true).Info("autorole removed")
			discordutil.RespondEphemeral(s, i, "Авто-роль удалена.")

		case command.AutoRoleSelectID:
			if len(data.Values) == 0 {
				return
			}
			roleID := data.Values[0]

			if err := setUC.Execute(ctx, i.GuildID, roleID); err != nil {
				log.WithError(err).Error("failed to set autorole")
				discordutil.RespondEphemeral(s, i, "Ошибка при установке авто-роли.")
				return
			}
			log.WithFields(logrus.Fields{"role_id": roleID, "notify": true}).Info("autorole set")

			g, err := getUC.Execute(ctx, i.GuildID)
			if err != nil || g == nil {
				discordutil.RespondEphemeral(s, i, fmt.Sprintf("Авто-роль установлена: <@&%s>.", roleID))
				return
			}

			if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseUpdateMessage,
				Data: &discordgo.InteractionResponseData{
					Content:    fmt.Sprintf("Текущая авто-роль: <@&%s>", g.AutoRoleID),
					Flags:      discordgo.MessageFlagsEphemeral,
					Components: []discordgo.MessageComponent{command.AutoRoleButtonsRow()},
				},
			}); err != nil {
				log.WithError(err).Error("autorole: failed to update message after role set")
			}
		}
	}
}
