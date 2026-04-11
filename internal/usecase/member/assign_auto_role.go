package member

import (
	"context"

	"github.com/ak1m1tsu/barman/internal/domain/guild"
)

// RoleAssigner assigns a Discord role to a guild member.
type RoleAssigner interface {
	AssignRole(ctx context.Context, guildID, userID, roleID string) error
}

// AssignAutoRoleUseCase assigns the configured auto-role when a member joins.
type AssignAutoRoleUseCase struct {
	repo     guild.Repository
	assigner RoleAssigner
}

func NewAssignAutoRole(repo guild.Repository, assigner RoleAssigner) *AssignAutoRoleUseCase {
	return &AssignAutoRoleUseCase{repo: repo, assigner: assigner}
}

func (uc *AssignAutoRoleUseCase) Execute(ctx context.Context, guildID, userID string) error {
	g, err := uc.repo.FindByID(ctx, guildID)
	if err != nil {
		return err
	}

	if g == nil || g.AutoRoleID == "" {
		return nil
	}

	return uc.assigner.AssignRole(ctx, guildID, userID, g.AutoRoleID)
}
