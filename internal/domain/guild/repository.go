package guild

import "context"

// Repository defines persistence operations for guild configuration.
type Repository interface {
	FindByID(ctx context.Context, guildID string) (*Guild, error)
	Save(ctx context.Context, guild *Guild) error
	Delete(ctx context.Context, guildID string) error
}
