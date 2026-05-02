package command

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	"github.com/ak1m1tsu/barman/internal/adapter/discordutil"
	cooldownuc "github.com/ak1m1tsu/barman/internal/usecase/cooldown"
	reactionuc "github.com/ak1m1tsu/barman/internal/usecase/reaction"
)

// ReactionMeta holds sentence templates for a reaction type.
type ReactionMeta struct {
	WithTarget    string // "%s обнимает %s"
	WithoutTarget string // "%s обнимает всех"
}

// ReactionOrder defines the canonical display order of reactions.
var ReactionOrder = []string{
	"hug", "pat", "kiss", "cuddle", "feed", "wave", "wink", "smile",
	"highfive", "handshake", "poke", "tickle", "lick", "bite", "slap", "punch",
	"love", "nuzzle", "shy", "nervous", "nosebleed", "brofist", "headbang", "sad", "peek",
}

// ReactionsMeta maps reaction type → sentence templates.
var ReactionsMeta = map[string]ReactionMeta{
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
}

// NewReactCommand returns the /react slash command and its handler.
// The handler fetches a GIF from the primary or fallback source, sends it as an embed,
// and triggers a reciprocal bot reaction when the target is the bot itself (subject to cooldown).
func NewReactCommand(fetchGIF *reactionuc.FetchGIFWithFallbackUseCase, checkAndSet *cooldownuc.CheckAndSetUseCase, incrementStat *reactionuc.IncrementStatUseCase, ownerIDs []string, timeout time.Duration) (*discordgo.ApplicationCommand, Handler) {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0, len(ReactionOrder))
	for _, key := range ReactionOrder {
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
			discordutil.RespondEphemeral(s, i, "Команда доступна только на сервере.")
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		opts := i.ApplicationCommandData().Options
		reactionType := opts[0].StringValue()
		meta := ReactionsMeta[reactionType]

		actor := discordutil.MemberDisplayName(i.Member)

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
					discordutil.RespondEphemeral(s, i, "Не удалось получить информацию о пользователе.")
					return
				}
				targetName = discordutil.MemberDisplayName(targetMember)
			}
		}

		sentence := discordutil.ReactionSentence(meta.WithTarget, meta.WithoutTarget, actor, targetName)

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

		gifURL, err := fetchGIF.Execute(ctx, reactionType)
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
				discordutil.RespondEphemeral(s, i, errMsg)
			}
			return
		}

		embed := discordutil.NewReactionEmbed(sentence, gifURL)

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

			respMsg, err := s.InteractionResponse(i.Interaction)
			if err != nil {
				log.WithError(err).Error("failed to fetch interaction response")
				return
			}

			botName := discordutil.MemberDisplayName(&discordgo.Member{User: s.State.User})
			botSentence := discordutil.ReactionSentence(meta.WithTarget, meta.WithoutTarget, botName, actor)
			if _, err := s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
				Reference: respMsg.Reference(),
				Embed:     discordutil.NewReactionEmbed(botSentence, botGIF),
			}); err != nil {
				log.WithError(err).Error("react: failed to send bot reaction")
			}
		}
	}

	return cmd, handler
}
