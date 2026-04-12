package handler

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	guilddomain "github.com/ak1m1tsu/barman/internal/domain/guild"
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
func NewMessageReactHandler(repo guilddomain.Repository, defaultPrefix string, fetchGIF *reactionuc.FetchGIFUseCase) func(*discordgo.Session, *discordgo.MessageCreate) {
	return func(s *discordgo.Session, msg *discordgo.MessageCreate) {
		if msg.Author == nil || msg.Author.Bot {
			return
		}

		prefix := defaultPrefix
		if g, err := repo.FindByID(context.Background(), msg.GuildID); err == nil && g != nil && g.Prefix != "" {
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
		gifURL, err := fetchGIF.Execute(context.Background(), reactionType)
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

		s.ChannelMessageSendComplex(msg.ChannelID, &discordgo.MessageSend{ //nolint:errcheck
			Embed: embed,
		})
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
