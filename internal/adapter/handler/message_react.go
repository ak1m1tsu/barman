package handler

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	"github.com/ak1m1tsu/barman/internal/adapter/command"
	"github.com/ak1m1tsu/barman/internal/adapter/discordutil"
	guilddomain "github.com/ak1m1tsu/barman/internal/domain/guild"
	cooldownuc "github.com/ak1m1tsu/barman/internal/usecase/cooldown"
	reactionuc "github.com/ak1m1tsu/barman/internal/usecase/reaction"
)

// NewMessageReactHandler handles prefix-based react commands.
// Supported formats:
//
//	!<type>           — no target
//	!<type> <@user>   — explicit mention target
//	!<type>           — when used as a reply, target is the replied-to author
//
// Priority: explicit mention > reply context > no target.
// The guild-specific prefix is fetched at runtime from repo; defaultPrefix is used as fallback.
// Pass a non-nil RateLimiter to enforce the same per-user cooldown as slash commands.
func NewMessageReactHandler(repo guilddomain.Repository, defaultPrefix string, rateLimiter *command.RateLimiter, fetchGIF *reactionuc.FetchGIFWithFallbackUseCase, checkAndSet *cooldownuc.CheckAndSetUseCase, incrementStat *reactionuc.IncrementStatUseCase, ownerIDs []string, timeout time.Duration) func(*discordgo.Session, *discordgo.MessageCreate) {
	return func(s *discordgo.Session, msg *discordgo.MessageCreate) {
		if msg.Author == nil || msg.Author.Bot {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
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
		meta, ok := command.ReactionsMeta[reactionType]
		if !ok {
			return
		}

		log := logrus.WithFields(logrus.Fields{
			"guild_id": msg.GuildID,
			"user_id":  msg.Author.ID,
			"reaction": reactionType,
			"command":  "react (prefix)",
			"notify":   true,
		})
		log.Info("command invoked")

		if rateLimiter != nil {
			if ok, remaining, violations := rateLimiter.Allow(msg.Author.ID, "react"); !ok {
				reply := command.RateLimitMessage(violations, remaining)
				if _, err := s.ChannelMessageSendReply(msg.ChannelID, reply, msg.Reference()); err != nil {
					log.WithError(err).Error("failed to send rate limit reply")
				}
				return
			}
		}

		// Resolve actor display name via guild member
		actorMember, err := s.GuildMember(msg.GuildID, msg.Author.ID)
		if err != nil {
			log.WithError(err).Error("failed to fetch actor guild member")
			return
		}
		actor := discordutil.MemberDisplayName(actorMember)

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
				targetName = discordutil.MemberDisplayName(targetMember)
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
						targetName = discordutil.MemberDisplayName(targetMember)
					}
				}
			}
		}

		sentence := discordutil.ReactionSentence(meta.WithTarget, meta.WithoutTarget, actor, targetName)

		gifURL, err := fetchGIF.Execute(ctx, reactionType)
		if err != nil {
			log.WithError(err).Error("failed to fetch reaction gif")
			if _, err := s.ChannelMessageSendReply(msg.ChannelID, "Не удалось получить GIF. Попробуйте позже.", msg.Reference()); err != nil {
				log.WithError(err).Error("react: failed to send error reply")
			}
			return
		}

		if _, err := s.ChannelMessageSendComplex(msg.ChannelID, &discordgo.MessageSend{
			Embed: discordutil.NewReactionEmbed(sentence, gifURL),
		}); err != nil {
			log.WithError(err).Error("failed to send reaction embed")
			return
		}

		if err := incrementStat.Execute(ctx, reactionType); err != nil {
			log.WithError(err).Warn("failed to increment reaction stat")
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

			botName := discordutil.MemberDisplayName(&discordgo.Member{User: s.State.User})
			botSentence := discordutil.ReactionSentence(meta.WithTarget, meta.WithoutTarget, botName, actor)
			if _, err := s.ChannelMessageSendComplex(msg.ChannelID, &discordgo.MessageSend{
				Reference: msg.Reference(),
				Embed:     discordutil.NewReactionEmbed(botSentence, botGIF),
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
