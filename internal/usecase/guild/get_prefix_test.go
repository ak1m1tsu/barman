package guild_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/ak1m1tsu/barman/internal/domain/guild"
	guilduc "github.com/ak1m1tsu/barman/internal/usecase/guild"
	mockguild "github.com/ak1m1tsu/barman/mocks/guild"
)

func TestGetPrefix_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("returns prefix when found", func(t *testing.T) {
		repo := mockguild.NewMockRepository(t)
		repo.EXPECT().FindByID(ctx, "guild1").Return(&domain.Guild{ID: "guild1", Prefix: "?"}, nil)

		uc := guilduc.NewGetPrefix(repo)
		prefix, err := uc.Execute(ctx, "guild1")
		require.NoError(t, err)
		assert.Equal(t, "?", prefix)
	})

	t.Run("returns empty string when guild not found", func(t *testing.T) {
		repo := mockguild.NewMockRepository(t)
		repo.EXPECT().FindByID(ctx, "guild1").Return(nil, nil)

		uc := guilduc.NewGetPrefix(repo)
		prefix, err := uc.Execute(ctx, "guild1")
		require.NoError(t, err)
		assert.Equal(t, "", prefix)
	})

	t.Run("returns empty string when prefix not set", func(t *testing.T) {
		repo := mockguild.NewMockRepository(t)
		repo.EXPECT().FindByID(ctx, "guild1").Return(&domain.Guild{ID: "guild1", Prefix: ""}, nil)

		uc := guilduc.NewGetPrefix(repo)
		prefix, err := uc.Execute(ctx, "guild1")
		require.NoError(t, err)
		assert.Equal(t, "", prefix)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repo := mockguild.NewMockRepository(t)
		repo.EXPECT().FindByID(ctx, "guild1").Return(nil, errors.New("db error"))

		uc := guilduc.NewGetPrefix(repo)
		_, err := uc.Execute(ctx, "guild1")
		assert.Error(t, err)
	})
}
