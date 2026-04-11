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

func TestGetAutoRole_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("returns guild when found", func(t *testing.T) {
		expected := &domain.Guild{ID: "guild1", AutoRoleID: "role1"}
		repo := mockguild.NewMockRepository(t)
		repo.EXPECT().FindByID(ctx, "guild1").Return(expected, nil)

		uc := guilduc.NewGetAutoRole(repo)
		g, err := uc.Execute(ctx, "guild1")
		require.NoError(t, err)
		assert.Equal(t, expected, g)
	})

	t.Run("returns nil when not found", func(t *testing.T) {
		repo := mockguild.NewMockRepository(t)
		repo.EXPECT().FindByID(ctx, "guild1").Return(nil, nil)

		uc := guilduc.NewGetAutoRole(repo)
		g, err := uc.Execute(ctx, "guild1")
		require.NoError(t, err)
		assert.Nil(t, g)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repo := mockguild.NewMockRepository(t)
		repo.EXPECT().FindByID(ctx, "guild1").Return(nil, errors.New("db error"))

		uc := guilduc.NewGetAutoRole(repo)
		_, err := uc.Execute(ctx, "guild1")
		assert.Error(t, err)
	})
}
