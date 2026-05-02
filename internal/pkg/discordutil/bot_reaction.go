package discordutil

import (
	"context"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	cooldownuc "github.com/ak1m1tsu/barman/internal/usecase/cooldown"
	reactionuc "github.com/ak1m1tsu/barman/internal/usecase/reaction"
)

// SendBotReactionReply sends a reciprocal bot reaction to channelID, referencing ref.
// It fetches a fresh GIF for reactionType, builds the sentence with botName→actor, and
// posts the embed. If the invoking user (userID) is an owner, the cooldown check is skipped;
// otherwise checkAndSet gates the reply. Returns without sending if the user is on cooldown.
func SendBotReactionReply(
	ctx context.Context,
	s *discordgo.Session,
	channelID string,
	ref *discordgo.MessageReference,
	reactionType string,
	withTarget string,
	withoutTarget string,
	botName string,
	actor string,
	fetchGIF *reactionuc.FetchGIFWithFallbackUseCase,
	checkAndSet *cooldownuc.CheckAndSetUseCase,
	ownerIDs []string,
	userID string,
	log *logrus.Entry,
) {
	isOwner := slices.Contains(ownerIDs, userID)
	if !isOwner {
		allowed, err := checkAndSet.Execute(ctx, userID)
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

	botSentence := ReactionSentence(withTarget, withoutTarget, botName, actor)
	if _, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Reference: ref,
		Embed:     NewReactionEmbed(botSentence, botGIF),
	}); err != nil {
		log.WithError(err).Error("failed to send bot reaction")
	}
}
