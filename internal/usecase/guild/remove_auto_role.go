package guild

import (
	"context"

	"github.com/ak1m1tsu/barman/internal/domain/guild"
)

// RemoveAutoRoleUseCase removes the auto-role configuration for a guild.
type RemoveAutoRoleUseCase struct {
	repo guild.Repository
}

func NewRemoveAutoRole(repo guild.Repository) *RemoveAutoRoleUseCase {
	return &RemoveAutoRoleUseCase{repo: repo}
}

func (uc *RemoveAutoRoleUseCase) Execute(ctx context.Context, guildID string) error {
	return uc.repo.Delete(ctx, guildID)
}
