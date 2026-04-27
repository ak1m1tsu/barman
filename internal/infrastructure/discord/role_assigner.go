package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// RoleAssigner implements usecase/member.RoleAssigner via the Discord API.
type RoleAssigner struct {
	session *discordgo.Session
}

// NewRoleAssigner returns a RoleAssigner that uses the given Discord session to assign roles.
func NewRoleAssigner(session *discordgo.Session) *RoleAssigner {
	return &RoleAssigner{session: session}
}

func (r *RoleAssigner) AssignRole(_ context.Context, guildID, userID, roleID string) error {
	return r.session.GuildMemberRoleAdd(guildID, userID, roleID)
}
