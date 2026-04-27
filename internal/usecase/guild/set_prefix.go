package guild

import (
	"context"

	"github.com/ak1m1tsu/barman/internal/domain/guild"
)

// SetPrefixUseCase sets the command prefix for a guild.
type SetPrefixUseCase struct {
	repo guild.Repository
}

// NewSetPrefix returns a SetPrefixUseCase backed by the given repository.
func NewSetPrefix(repo guild.Repository) *SetPrefixUseCase {
	return &SetPrefixUseCase{repo: repo}
}

func (uc *SetPrefixUseCase) Execute(ctx context.Context, guildID, prefix string) error {
	return uc.repo.Save(ctx, &guild.Guild{
		ID:     guildID,
		Prefix: prefix,
	})
}
