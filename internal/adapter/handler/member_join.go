package handler

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	memberuc "github.com/ak1m1tsu/barman/internal/usecase/member"
)

// NewMemberJoinHandler returns a GuildMemberAdd event handler that assigns
// the configured auto-role to each new member.
func NewMemberJoinHandler(uc *memberuc.AssignAutoRoleUseCase) func(*discordgo.Session, *discordgo.GuildMemberAdd) {
	return func(_ *discordgo.Session, e *discordgo.GuildMemberAdd) {
		log := logrus.WithFields(logrus.Fields{
			"guild_id": e.GuildID,
			"user_id":  e.User.ID,
		})
		if err := uc.Execute(context.Background(), e.GuildID, e.User.ID); err != nil {
			log.WithError(err).Error("autorole: failed to assign role")
			return
		}
		log.Info("autorole: role assigned")
	}
}
