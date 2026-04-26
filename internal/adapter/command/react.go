package command

import (
	"context"
	"fmt"
	"math/rand"
	"slices"
	"time"

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
	"love", "nuzzle", "shy", "nervous", "nosebleed", "brofist", "headbang", "sad", "peek",
	"myatniy",
}

// nsfwReactions lists reaction types that require an age-restricted (NSFW) channel.
var nsfwReactions = map[string]bool{
	"myatniy": true,
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
	"love":      {"%s любит %s", "%s любит всех"},
	"nuzzle":    {"%s тыкается носом в %s", "%s тыкается носом"},
	"shy":       {"%s стесняется перед %s", "%s стесняется"},
	"nervous":   {"%s нервничает перед %s", "%s нервничает"},
	"nosebleed": {"%s краснеет от %s", "%s краснеет"},
	"brofist":   {"%s бьёт кулаком %s", "%s бьёт кулаком"},
	"headbang":  {"%s хэдбэнгит с %s", "%s хэдбэнгит"},
	"sad":       {"%s грустит с %s", "%s грустит"},
	"peek":      {"%s подглядывает за %s", "%s подглядывает"},
	"myatniy":   {"%s делает мятный %s", "%s делает мятный всем"},
}

func NewReactCommand(fetchGIF *reactionuc.FetchGIFWithFallbackUseCase, nsfwFetchGIF *reactionuc.FetchGIFWithFallbackUseCase, checkAndSet *cooldownuc.CheckAndSetUseCase, incrementStat *reactionuc.IncrementStatUseCase, ownerIDs []string, nsfwAllowedUsers map[string][]string, timeout time.Duration) (*discordgo.ApplicationCommand, Handler) {
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

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		opts := i.ApplicationCommandData().Options
		reactionType := opts[0].StringValue()
		meta := reactionsMeta[reactionType]

		isOwner := slices.Contains(ownerIDs, i.Member.User.ID)
		if nsfwReactions[reactionType] && !isOwner {
			if _, ok := nsfwAllowedUsers[i.Member.User.ID]; !ok {
				respondEphemeral(s, i, "У тебя нет доступа к этой реакции.")
				return
			}
		}

		actor := memberDisplayName(i.Member)

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
				targetName = memberDisplayName(targetMember)
			}
		}

		if nsfwReactions[reactionType] {
			if !isOwner {
				allowedTargets := nsfwAllowedUsers[i.Member.User.ID]
				if targetID != "" && !slices.Contains(allowedTargets, targetID) {
					respondEphemeral(s, i, "Этот пользователь не может быть целью этой реакции.")
					return
				}
			}
			ch, err := s.Channel(i.ChannelID)
			if err != nil || !ch.NSFW {
				respondEphemeral(s, i, "Эта реакция доступна только в NSFW-каналах.")
				return
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

		gif := fetchGIF
		if nsfwReactions[reactionType] {
			gif = nsfwFetchGIF
		}
		gifURL, err := gif.Execute(ctx, reactionType)
		if err != nil {
			log.WithError(err).Error("failed to fetch reaction gif")
			errMsg := "Не удалось получить GIF. Попробуйте позже."
			if pingTarget {
				if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &errMsg,
				}); err != nil {
					log.WithError(err).Error("react: failed to edit response with error message")
				}
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
			if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &empty,
				Embeds:  &[]*discordgo.MessageEmbed{embed},
			}); err != nil {
				log.WithError(err).Error("react: failed to edit response with embed")
			}
		} else {
			if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
				},
			}); err != nil {
				log.WithError(err).Error("react: failed to send response")
			}
		}

		if err := incrementStat.Execute(ctx, reactionType); err != nil {
			log.WithError(err).Warn("failed to increment reaction stat")
		}

		// If the target is the bot — respond with the same reaction back.
		botID := s.State.User.ID
		if targetID == botID {
			isOwner := slices.Contains(ownerIDs, i.Member.User.ID)
			if !isOwner {
				allowed, err := checkAndSet.Execute(ctx, i.Member.User.ID)
				if err != nil {
					log.WithError(err).Error("failed to check reaction cooldown")
					return
				}
				if !allowed {
					return
				}
			}

			botGIF, err := fetchGIF.Execute(ctx, reactionType)
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

			botName := memberDisplayName(&discordgo.Member{User: s.State.User})
			botSentence := fmt.Sprintf(meta.withTarget, botName, actor)
			if _, err := s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
				Reference: respMsg.Reference(),
				Embed: &discordgo.MessageEmbed{
					Title: botSentence,
					Color: rand.Intn(0xFFFFFF + 1),
					Image: &discordgo.MessageEmbedImage{URL: botGIF},
				},
			}); err != nil {
				log.WithError(err).Error("react: failed to send bot reaction")
			}
		}
	}

	return cmd, handler
}

// memberDisplayName returns the best available display name for a guild member.
func memberDisplayName(m *discordgo.Member) string {
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
