package guild

import (
	"context"

	"github.com/ak1m1tsu/barman/internal/domain/guild"
)

// RemovePrefixUseCase resets the command prefix for a guild to the global default.
type RemovePrefixUseCase struct {
	repo guild.Repository
}

// NewRemovePrefix returns a RemovePrefixUseCase backed by the given repository.
func NewRemovePrefix(repo guild.Repository) *RemovePrefixUseCase {
	return &RemovePrefixUseCase{repo: repo}
}

func (uc *RemovePrefixUseCase) Execute(ctx context.Context, guildID string) error {
	g, err := uc.repo.FindByID(ctx, guildID)
	if err != nil {
		return err
	}
	if g == nil {
		return nil
	}
	g.Prefix = ""
	return uc.repo.Save(ctx, g)
}
