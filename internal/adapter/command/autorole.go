package command

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"

	guilduc "github.com/ak1m1tsu/barman/internal/usecase/guild"
)

func NewAutoRoleCommand(
	setUC *guilduc.SetAutoRoleUseCase,
	getUC *guilduc.GetAutoRoleUseCase,
	removeUC *guilduc.RemoveAutoRoleUseCase,
) (*discordgo.ApplicationCommand, Handler) {
	cmd := &discordgo.ApplicationCommand{
		Name:        "autorole",
		Description: "Управление авто-ролью сервера",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "set",
				Description: "Установить авто-роль",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionRole,
						Name:        "role",
						Description: "Роль для выдачи новым участникам",
						Required:    true,
					},
				},
			},
			{
				Name:        "remove",
				Description: "Удалить авто-роль",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
			{
				Name:        "info",
				Description: "Показать текущую авто-роль",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
		},
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

		subCmd := i.ApplicationCommandData().Options[0]
		ctx := context.Background()

		switch subCmd.Name {
		case "set":
			role := subCmd.Options[0].RoleValue(s, i.GuildID)
			if err := setUC.Execute(ctx, i.GuildID, role.ID); err != nil {
				respond(s, i, "Ошибка при установке авто-роли.")
				return
			}
			respond(s, i, fmt.Sprintf("Авто-роль установлена: <@&%s>", role.ID))

		case "remove":
			if err := removeUC.Execute(ctx, i.GuildID); err != nil {
				respond(s, i, "Ошибка при удалении авто-роли.")
				return
			}
			respond(s, i, "Авто-роль удалена.")

		case "info":
			g, err := getUC.Execute(ctx, i.GuildID)
			if err != nil || g == nil || g.AutoRoleID == "" {
				respond(s, i, "Авто-роль не установлена.")
				return
			}
			respond(s, i, fmt.Sprintf("Текущая авто-роль: <@&%s>", g.AutoRoleID))
		}
	}

	return cmd, handler
}

func respond(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{ //nolint:errcheck
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: content},
	})
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
