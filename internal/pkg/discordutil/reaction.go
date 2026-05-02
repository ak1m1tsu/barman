package discordutil

import (
	"fmt"
	"math/rand"

	"github.com/bwmarrin/discordgo"
)

// MemberDisplayName returns the best available display name for a guild member:
// server nickname > global display name > username. Returns "кто-то" when the
// member has no user attached (e.g. a synthetic member object).
func MemberDisplayName(m *discordgo.Member) string {
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

// ReactionSentence builds the display sentence for a reaction.
// If target is empty, withoutTarget is used (formatted with actor only);
// otherwise withTarget is used (formatted with actor and target).
func ReactionSentence(withTarget, withoutTarget, actor, target string) string {
	if target == "" {
		return fmt.Sprintf(withoutTarget, actor)
	}
	return fmt.Sprintf(withTarget, actor, target)
}

// NewReactionEmbed builds a Discord embed for an anime reaction GIF with a random colour.
func NewReactionEmbed(title, gifURL string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: title,
		Color: rand.Intn(0xFFFFFF + 1),
		Image: &discordgo.MessageEmbedImage{URL: gifURL},
	}
}
