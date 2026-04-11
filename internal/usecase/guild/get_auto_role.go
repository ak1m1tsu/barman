package guild

import (
	"context"

	"github.com/ak1m1tsu/barman/internal/domain/guild"
)

// GetAutoRoleUseCase retrieves the auto-role configuration for a guild.
type GetAutoRoleUseCase struct {
	repo guild.Repository
}

func NewGetAutoRole(repo guild.Repository) *GetAutoRoleUseCase {
	return &GetAutoRoleUseCase{repo: repo}
}

func (uc *GetAutoRoleUseCase) Execute(ctx context.Context, guildID string) (*guild.Guild, error) {
	return uc.repo.FindByID(ctx, guildID)
}
