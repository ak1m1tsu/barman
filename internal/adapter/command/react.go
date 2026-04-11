package command

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	reactionuc "github.com/ak1m1tsu/barman/internal/usecase/reaction"
)

type reactionMeta struct {
	withTarget    string // format string: "%s обнимает %s"
	withoutTarget string // format string: "%s обнимает всех"
	color         int
}

var reactionOrder = []string{
	"hug", "pat", "kiss", "cuddle", "feed", "wave", "wink", "smile",
	"highfive", "handshake", "poke", "tickle", "lick", "bite", "slap", "punch",
}

var reactionsMeta = map[string]reactionMeta{
	"hug":       {"%s обнимает %s", "%s обнимает всех", 0xFF6B9D},
	"pat":       {"%s гладит %s", "%s гладит всех", 0xFFB347},
	"kiss":      {"%s целует %s", "%s целует всех", 0xFF85A1},
	"cuddle":    {"%s прижимается к %s", "%s прижимается ко всем", 0xFFB6C1},
	"feed":      {"%s кормит %s", "%s кормит всех", 0x98D8C8},
	"wave":      {"%s машет %s", "%s машет всем", 0x87CEEB},
	"wink":      {"%s подмигивает %s", "%s подмигивает всем", 0xB0E0E6},
	"smile":     {"%s улыбается %s", "%s улыбается всем", 0xFFF44F},
	"highfive":  {"%s даёт пять %s", "%s даёт пять всем", 0xFFD700},
	"handshake": {"%s жмёт руку %s", "%s жмёт руку всем", 0x98FB98},
	"poke":      {"%s тыкает %s", "%s тыкает всех", 0xA8D8EA},
	"tickle":    {"%s щекочет %s", "%s щекочет всех", 0xFFE4B5},
	"lick":      {"%s лижет %s", "%s лижет всех", 0xDDA0DD},
	"bite":      {"%s кусает %s", "%s кусает всех", 0xFF6633},
	"slap":      {"%s даёт пощёчину %s", "%s даёт пощёчину всем", 0xFF4444},
	"punch":     {"%s бьёт %s", "%s бьёт всех", 0xFF4444},
}

func NewReactCommand(fetchGIF *reactionuc.FetchGIFUseCase) (*discordgo.ApplicationCommand, Handler) {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0, len(reactionOrder))
	for _, key := range reactionOrder {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  key,
			Value: key,
		})
	}

	cmd := &discordgo.ApplicationCommand{
		Name:        "react",
		Description: "Отправить аниме-реакцию",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "type",
				Description: "Тип реакции",
				Required:    true,
				Choices:     choices,
			},
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "Цель реакции (по умолчанию — все)",
				Required:    false,
			},
		},
	}

	handler := func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Member == nil {
			respondEphemeral(s, i, "Команда доступна только на сервере.")
			return
		}

		opts := i.ApplicationCommandData().Options
		reactionType := opts[0].StringValue()
		meta := reactionsMeta[reactionType]

		actorName := displayName(i.Member)

		var sentence string
		if len(opts) > 1 {
			targetUser := opts[1].UserValue(s)
			if targetUser.ID == i.Member.User.ID {
				sentence = fmt.Sprintf(meta.withTarget, actorName, "себя")
			} else {
				targetMember, err := s.GuildMember(i.GuildID, targetUser.ID)
				if err != nil {
					logrus.WithError(err).WithFields(logrus.Fields{
						"guild_id": i.GuildID,
						"user_id":  targetUser.ID,
						"command":  "react " + reactionType,
					}).Error("failed to fetch target guild member")
					respondEphemeral(s, i, "Не удалось получить информацию о пользователе.")
					return
				}
				sentence = fmt.Sprintf(meta.withTarget, actorName, displayName(targetMember))
			}
		} else {
			sentence = fmt.Sprintf(meta.withoutTarget, actorName)
		}

		gifURL, err := fetchGIF.Execute(context.Background(), reactionType)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"guild_id": i.GuildID,
				"reaction": reactionType,
				"command":  "react",
			}).Error("failed to fetch reaction gif")
			respondEphemeral(s, i, "Не удалось получить GIF. Попробуйте позже.")
			return
		}

		embed := &discordgo.MessageEmbed{
			Title: sentence,
			Color: meta.color,
			Image: &discordgo.MessageEmbedImage{URL: gifURL},
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

// displayName returns the best available display name for a guild member.
func displayName(m *discordgo.Member) string {
	if m.Nick != "" {
		return m.Nick
	}
	if m.User != nil {
		if m.User.GlobalName != "" {
			return m.User.GlobalName
		}
		return m.User.Username
	}
	return "кто-то"
}
