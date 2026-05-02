package command

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	"github.com/ak1m1tsu/barman/internal/pkg/discordutil"
)

// NewUserInfoCommand returns the /userinfo slash command and its handler.
// When called without an argument the handler shows information about the invoking member.
func NewUserInfoCommand() (*discordgo.ApplicationCommand, Handler) {
	cmd := &discordgo.ApplicationCommand{
		Name:        "userinfo",
		Description: "Информация о пользователе",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "Пользователь (по умолчанию — вы)",
				Required:    false,
			},
		},
	}

	handler := func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Member == nil {
			discordutil.RespondEphemeral(s, i, "Команда доступна только на сервере.")
			return
		}

		var target *discordgo.User
		if opts := i.ApplicationCommandData().Options; len(opts) > 0 {
			target = opts[0].UserValue(s)
		} else {
			target = i.Member.User
		}

		member, err := s.GuildMember(i.GuildID, target.ID)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"guild_id": i.GuildID,
				"user_id":  target.ID,
				"command":  "userinfo",
			}).Error("failed to fetch guild member")
			discordutil.RespondEphemeral(s, i, "Не удалось получить информацию о пользователе.")
			return
		}

		roles := "нет"
		if len(member.Roles) > 0 {
			roles = ""
			for _, r := range member.Roles {
				roles += fmt.Sprintf("<@&%s> ", r)
			}
		}

		joinedAt := "неизвестно"
		if !member.JoinedAt.IsZero() {
			joinedAt = member.JoinedAt.Format("02.01.2006")
		}

		createdAt := snowflakeToTime(target.ID).Format("02.01.2006")

		embed := &discordgo.MessageEmbed{
			Title: target.Username,
			Color: ColorDiscordBranding,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: target.AvatarURL("256"),
			},
			Fields: []*discordgo.MessageEmbedField{
				{Name: "ID", Value: target.ID, Inline: true},
				{Name: "Аккаунт создан", Value: createdAt, Inline: true},
				{Name: "Вступил на сервер", Value: joinedAt, Inline: true},
				{Name: "Роли", Value: roles},
			},
		}

		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
				Flags:  discordgo.MessageFlagsEphemeral,
			},
		}); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"guild_id": i.GuildID,
				"user_id":  target.ID,
			}).Error("userinfo: failed to send response")
		}
	}

	return cmd, handler
}

// snowflakeToTime converts a Discord snowflake ID to a UTC time.
func snowflakeToTime(id string) time.Time {
	snowflake, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return time.Time{}
	}
	const discordEpoch = int64(1420070400000)
	ms := (snowflake >> 22) + discordEpoch
	return time.Unix(ms/1000, (ms%1000)*int64(time.Millisecond)).UTC()
}
