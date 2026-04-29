package guild

import (
	"context"

	"github.com/ak1m1tsu/barman/internal/domain/guild"
)

// SetAutoRoleUseCase sets the auto-role for a guild.
type SetAutoRoleUseCase struct {
	repo guild.Repository
}

// NewSetAutoRole returns a SetAutoRoleUseCase backed by the given repository.
func NewSetAutoRole(repo guild.Repository) *SetAutoRoleUseCase {
	return &SetAutoRoleUseCase{repo: repo}
}

func (uc *SetAutoRoleUseCase) Execute(ctx context.Context, guildID, roleID string) error {
	g, err := uc.repo.FindByID(ctx, guildID)
	if err != nil {
		return err
	}
	if g == nil {
		g = &guild.Guild{ID: guildID}
	}
	g.AutoRoleID = roleID
	return uc.repo.Save(ctx, g)
}
