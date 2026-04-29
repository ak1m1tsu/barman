package guild

import (
	"context"

	"github.com/ak1m1tsu/barman/internal/domain/guild"
)

// RemoveAutoRoleUseCase removes the auto-role configuration for a guild.
type RemoveAutoRoleUseCase struct {
	repo guild.Repository
}

// NewRemoveAutoRole returns a RemoveAutoRoleUseCase backed by the given repository.
func NewRemoveAutoRole(repo guild.Repository) *RemoveAutoRoleUseCase {
	return &RemoveAutoRoleUseCase{repo: repo}
}

func (uc *RemoveAutoRoleUseCase) Execute(ctx context.Context, guildID string) error {
	g, err := uc.repo.FindByID(ctx, guildID)
	if err != nil {
		return err
	}
	if g == nil {
		return nil
	}
	g.AutoRoleID = ""
	return uc.repo.Save(ctx, g)
}
