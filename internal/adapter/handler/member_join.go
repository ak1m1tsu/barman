package handler

import (
	"context"
	"log"

	"github.com/bwmarrin/discordgo"

	memberuc "github.com/ak1m1tsu/barman/internal/usecase/member"
)

// NewMemberJoinHandler returns a GuildMemberAdd event handler that assigns
// the configured auto-role to each new member.
func NewMemberJoinHandler(uc *memberuc.AssignAutoRoleUseCase) func(*discordgo.Session, *discordgo.GuildMemberAdd) {
	return func(_ *discordgo.Session, e *discordgo.GuildMemberAdd) {
		if err := uc.Execute(context.Background(), e.GuildID, e.User.ID); err != nil {
			log.Printf("autorole: failed to assign role to %s in %s: %v", e.User.ID, e.GuildID, err)
		}
	}
}
