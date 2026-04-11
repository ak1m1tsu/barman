package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	reactionuc "github.com/ak1m1tsu/barman/internal/usecase/reaction"
)

type msgReactionMeta struct {
	withTarget    string
	withoutTarget string
	color         int
}

var msgReactions = map[string]msgReactionMeta{
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

// NewMessageReactHandler handles prefix-based react commands.
// Supported formats:
//
//	!<type>           — no target
//	!<type> <@user>   — explicit mention target
//	!<type>           — when used as a reply, target is the replied-to author
//
// Priority: explicit mention > reply context > no target.
func NewMessageReactHandler(prefix string, fetchGIF *reactionuc.FetchGIFUseCase) func(*discordgo.Session, *discordgo.MessageCreate) {
	return func(s *discordgo.Session, msg *discordgo.MessageCreate) {
		if msg.Author == nil || msg.Author.Bot {
			return
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
			Color: meta.color,
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
