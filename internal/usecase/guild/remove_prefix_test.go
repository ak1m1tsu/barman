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

func TestRemovePrefix_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("clears prefix while preserving auto_role", func(t *testing.T) {
		existing := &domain.Guild{ID: "guild1", AutoRoleID: "role1", Prefix: "!"}
		repo := mockguild.NewMockRepository(t)
		repo.EXPECT().FindByID(ctx, "guild1").Return(existing, nil)
		repo.EXPECT().Save(ctx, &domain.Guild{ID: "guild1", AutoRoleID: "role1", Prefix: ""}).Return(nil)

		uc := guilduc.NewRemovePrefix(repo)
		require.NoError(t, uc.Execute(ctx, "guild1"))
	})

	t.Run("no-op when guild not found", func(t *testing.T) {
		repo := mockguild.NewMockRepository(t)
		repo.EXPECT().FindByID(ctx, "guild1").Return(nil, nil)

		uc := guilduc.NewRemovePrefix(repo)
		require.NoError(t, uc.Execute(ctx, "guild1"))
	})

	t.Run("propagates FindByID error", func(t *testing.T) {
		repo := mockguild.NewMockRepository(t)
		repo.EXPECT().FindByID(ctx, "guild1").Return(nil, errors.New("db error"))

		uc := guilduc.NewRemovePrefix(repo)
		assert.Error(t, uc.Execute(ctx, "guild1"))
	})

	t.Run("propagates Save error", func(t *testing.T) {
		existing := &domain.Guild{ID: "guild1", Prefix: "!"}
		repo := mockguild.NewMockRepository(t)
		repo.EXPECT().FindByID(ctx, "guild1").Return(existing, nil)
		repo.EXPECT().Save(ctx, &domain.Guild{ID: "guild1", Prefix: ""}).
			Return(errors.New("db error"))

		uc := guilduc.NewRemovePrefix(repo)
		assert.Error(t, uc.Execute(ctx, "guild1"))
	})
}
