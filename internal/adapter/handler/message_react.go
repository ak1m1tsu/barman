package handler

import (
	"context"
	"fmt"
	"math/rand"
	"slices"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	guilddomain "github.com/ak1m1tsu/barman/internal/domain/guild"
	cooldownuc "github.com/ak1m1tsu/barman/internal/usecase/cooldown"
	reactionuc "github.com/ak1m1tsu/barman/internal/usecase/reaction"
)

type msgReactionMeta struct {
	withTarget    string
	withoutTarget string
}

var msgReactions = map[string]msgReactionMeta{
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

// NewMessageReactHandler handles prefix-based react commands.
// Supported formats:
//
//	!<type>           — no target
//	!<type> <@user>   — explicit mention target
//	!<type>           — when used as a reply, target is the replied-to author
//
// Priority: explicit mention > reply context > no target.
// The guild-specific prefix is fetched at runtime from repo; defaultPrefix is used as fallback.
func NewMessageReactHandler(repo guilddomain.Repository, defaultPrefix string, fetchGIF *reactionuc.FetchGIFWithFallbackUseCase, checkAndSet *cooldownuc.CheckAndSetUseCase, ownerIDs []string) func(*discordgo.Session, *discordgo.MessageCreate) {
	return func(s *discordgo.Session, msg *discordgo.MessageCreate) {
		if msg.Author == nil || msg.Author.Bot {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		prefix := defaultPrefix
		if g, err := repo.FindByID(ctx, msg.GuildID); err == nil && g != nil && g.Prefix != "" {
			prefix = g.Prefix
		}

		if !strings.HasPrefix(msg.Content, prefix) {
			return
		}

		remainder := strings.TrimSpace(strings.TrimPrefix(msg.Content, prefix))
		parts := strings.SplitN(remainder, " ", 2)

		reactionType := parts[0]
		meta, ok := msgReactions[reactionType]
		if !ok {
			return
		}

		log := logrus.WithFields(logrus.Fields{
			"guild_id": msg.GuildID,
			"user_id":  msg.Author.ID,
			"reaction": reactionType,
			"command":  "react (prefix)",
		})
		log.Info("command invoked")

		// Resolve actor display name via guild member
		actorMember, err := s.GuildMember(msg.GuildID, msg.Author.ID)
		if err != nil {
			log.WithError(err).Error("failed to fetch actor guild member")
			return
		}
		actor := memberDisplayName(actorMember)

		// Resolve target: explicit mention > reply context > none
		var targetID string
		var targetName string

		if len(parts) == 2 {
			targetID = parseMention(parts[1])
		}

		if targetID != "" {
			// Explicit mention
			if targetID == msg.Author.ID {
				targetName = "себя"
			} else {
				targetMember, err := s.GuildMember(msg.GuildID, targetID)
				if err != nil {
					log.WithError(err).Error("failed to fetch target guild member")
					return
				}
				targetName = memberDisplayName(targetMember)
			}
		} else if msg.MessageReference != nil {
			// Fall back to reply context
			refMsg, err := s.ChannelMessage(msg.ChannelID, msg.MessageReference.MessageID)
			if err != nil {
				log.WithError(err).Error("failed to fetch referenced message")
			} else if refMsg.Author != nil {
				targetID = refMsg.Author.ID
				if targetID == msg.Author.ID {
					targetName = "себя"
				} else {
					targetMember, err := s.GuildMember(msg.GuildID, targetID)
					if err != nil {
						targetName = refMsg.Author.Username
					} else {
						targetName = memberDisplayName(targetMember)
					}
				}
			}
		}

		// Build sentence
		var sentence string
		if targetName == "" {
			sentence = fmt.Sprintf(meta.withoutTarget, actor)
		} else {
			sentence = fmt.Sprintf(meta.withTarget, actor, targetName)
		}

		// Fetch GIF
		gifURL, err := fetchGIF.Execute(ctx, reactionType)
		if err != nil {
			log.WithError(err).Error("failed to fetch reaction gif")
			s.ChannelMessageSendReply(msg.ChannelID, "Не удалось получить GIF. Попробуйте позже.", msg.Reference()) //nolint:errcheck
			return
		}

		embed := &discordgo.MessageEmbed{
			Title: sentence,
			Color: rand.Intn(0xFFFFFF + 1),
			Image: &discordgo.MessageEmbedImage{URL: gifURL},
		}

		if _, err := s.ChannelMessageSendComplex(msg.ChannelID, &discordgo.MessageSend{
			Embed: embed,
		}); err != nil {
			log.WithError(err).Error("failed to send reaction embed")
			return
		}

		// If the target is the bot — respond with the same reaction back.
		botID := s.State.User.ID
		if targetID == botID {
			isOwner := slices.Contains(ownerIDs, msg.Author.ID)
			if !isOwner {
				allowed, err := checkAndSet.Execute(ctx, msg.Author.ID)
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

			botName := memberDisplayName(&discordgo.Member{User: s.State.User})
			botSentence := fmt.Sprintf(meta.withTarget, botName, actor)
			if _, err := s.ChannelMessageSendComplex(msg.ChannelID, &discordgo.MessageSend{
				Reference: msg.Reference(),
				Embed: &discordgo.MessageEmbed{
					Title: botSentence,
					Color: rand.Intn(0xFFFFFF + 1),
					Image: &discordgo.MessageEmbedImage{URL: botGIF},
				},
			}); err != nil {
				log.WithError(err).Error("failed to send bot reaction message")
			}
		}
	}
}

// parseMention extracts a user ID from a Discord mention string (<@userId> or <@!userId>).
// Returns an empty string if s is not a valid mention.
func parseMention(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "<@") || !strings.HasSuffix(s, ">") {
		return ""
	}
	s = strings.TrimPrefix(s, "<@")
	s = strings.TrimPrefix(s, "!")
	s = strings.TrimSuffix(s, ">")
	return s
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
