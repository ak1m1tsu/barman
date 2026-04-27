package guild

import (
	"context"

	"github.com/ak1m1tsu/barman/internal/domain/guild"
)

// GetPrefixUseCase retrieves the command prefix for a guild.
// Returns an empty string when no custom prefix is configured.
type GetPrefixUseCase struct {
	repo guild.Repository
}

// NewGetPrefix returns a GetPrefixUseCase backed by the given repository.
func NewGetPrefix(repo guild.Repository) *GetPrefixUseCase {
	return &GetPrefixUseCase{repo: repo}
}

func (uc *GetPrefixUseCase) Execute(ctx context.Context, guildID string) (string, error) {
	g, err := uc.repo.FindByID(ctx, guildID)
	if err != nil {
		return "", err
	}
	if g == nil {
		return "", nil
	}
	return g.Prefix, nil
}
