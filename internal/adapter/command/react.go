package command

import (
	"context"
	"fmt"
	"math/rand"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	cooldownuc "github.com/ak1m1tsu/barman/internal/usecase/cooldown"
	reactionuc "github.com/ak1m1tsu/barman/internal/usecase/reaction"
)

type reactionMeta struct {
	withTarget    string // "%s обнимает %s"
	withoutTarget string // "%s обнимает всех"
}

var reactionOrder = []string{
	"hug", "pat", "kiss", "cuddle", "feed", "wave", "wink", "smile",
	"highfive", "handshake", "poke", "tickle", "lick", "bite", "slap", "punch",
}

var reactionsMeta = map[string]reactionMeta{
	"hug":       {"%s обнимает %s", "%s обнимает всех"},
	"pat":       {"%s гладит %s", "%s гладит всех"},
	"kiss":      {"%s целует %s", "%s целует всех"},
	"cuddle":    {"%s прижимается к %s", "%s прижимается ко всем"},
	"feed":      {"%s кормит %s", "%s кормит всех"},
	"wave":      {"%s машет %s", "%s машет всем"},
	"wink":      {"%s подмигивает %s", "%s подмигивает всем"},
	"smile":     {"%s улыбается %s", "%s улыбается всем"},
	"highfive":  {"%s даёт пять %s", "%s даёт пять всем"},
	"handshake": {"%s жмёт руку %s", "%s жмёт руку всем"},
	"poke":      {"%s тыкает %s", "%s тыкает всех"},
	"tickle":    {"%s щекочет %s", "%s щекочет всех"},
	"lick":      {"%s лижет %s", "%s лижет всех"},
	"bite":      {"%s кусает %s", "%s кусает всех"},
	"slap":      {"%s даёт пощёчину %s", "%s даёт пощёчину всем"},
	"punch":     {"%s бьёт %s", "%s бьёт всех"},
}

func NewReactCommand(fetchGIF *reactionuc.FetchGIFUseCase, checkAndSet *cooldownuc.CheckAndSetUseCase, ownerIDs []string) (*discordgo.ApplicationCommand, Handler) {
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

		actor := displayName(i.Member)

		// Resolve target
		var targetID string
		var targetName string
		if len(opts) > 1 {
			targetUser := opts[1].UserValue(s)
			targetID = targetUser.ID
			if targetID == i.Member.User.ID {
				targetName = "себя"
			} else {
				targetMember, err := s.GuildMember(i.GuildID, targetID)
				if err != nil {
					logrus.WithError(err).WithFields(logrus.Fields{
						"guild_id": i.GuildID,
						"user_id":  targetID,
						"command":  "react " + reactionType,
					}).Error("failed to fetch target guild member")
					respondEphemeral(s, i, "Не удалось получить информацию о пользователе.")
					return
				}
				targetName = displayName(targetMember)
			}
		}

		// Build sentence
		var sentence string
		if targetName == "" {
			sentence = fmt.Sprintf(meta.withoutTarget, actor)
		} else {
			sentence = fmt.Sprintf(meta.withTarget, actor, targetName)
		}

		log := logrus.WithFields(logrus.Fields{
			"guild_id": i.GuildID,
			"user_id":  i.Member.User.ID,
			"reaction": reactionType,
			"command":  "react",
		})

		// When targeting another user: ping them first, then edit to embed.
		// This ensures the target receives a notification.
		pingTarget := targetID != "" && targetID != i.Member.User.ID
		if pingTarget {
			if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("<@%s>", targetID),
					AllowedMentions: &discordgo.MessageAllowedMentions{
						Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeUsers},
					},
				},
			}); err != nil {
				log.WithError(err).Error("failed to send ping")
				return
			}
		}

		gifURL, err := fetchGIF.Execute(context.Background(), reactionType)
		if err != nil {
			log.WithError(err).Error("failed to fetch reaction gif")
			errMsg := "Не удалось получить GIF. Попробуйте позже."
			if pingTarget {
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{ //nolint:errcheck
					Content: &errMsg,
				})
			} else {
				respondEphemeral(s, i, errMsg)
			}
			return
		}

		embed := &discordgo.MessageEmbed{
			Title: sentence,
			Color: rand.Intn(0xFFFFFF + 1),
			Image: &discordgo.MessageEmbedImage{URL: gifURL},
		}

		if pingTarget {
			empty := ""
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{ //nolint:errcheck
				Content: &empty,
				Embeds:  &[]*discordgo.MessageEmbed{embed},
			})
		} else {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{ //nolint:errcheck
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
				},
			})
		}

		// If the target is the bot — respond with the same reaction back.
		botID := s.State.User.ID
		if targetID == botID {
			isOwner := slices.Contains(ownerIDs, i.Member.User.ID)
			if !isOwner {
				allowed, err := checkAndSet.Execute(context.Background(), i.Member.User.ID)
				if err != nil {
					log.WithError(err).Error("failed to check reaction cooldown")
					return
				}
				if !allowed {
					return
				}
			}

			botGIF, err := fetchGIF.Execute(context.Background(), reactionType)
			if err != nil {
				log.WithError(err).Error("failed to fetch bot reaction gif")
				return
			}

			// Fetch the original interaction response to reply to it.
			respMsg, err := s.InteractionResponse(i.Interaction)
			if err != nil {
				log.WithError(err).Error("failed to fetch interaction response")
				return
			}

			botName := displayName(&discordgo.Member{User: s.State.User})
			botSentence := fmt.Sprintf(meta.withTarget, botName, actor)
			s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{ //nolint:errcheck
				Reference: respMsg.Reference(),
				Content:   fmt.Sprintf("<@%s>", i.Member.User.ID),
				Embed: &discordgo.MessageEmbed{
					Title: botSentence,
					Color: rand.Intn(0xFFFFFF + 1),
					Image: &discordgo.MessageEmbedImage{URL: botGIF},
				},
			})
		}
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
