package guild_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	guilduc "github.com/ak1m1tsu/barman/internal/usecase/guild"
	mockguild "github.com/ak1m1tsu/barman/mocks/guild"
)

func TestRemoveAutoRole_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes guild settings", func(t *testing.T) {
		repo := mockguild.NewMockRepository(t)
		repo.EXPECT().Delete(ctx, "guild1").Return(nil)

		uc := guilduc.NewRemoveAutoRole(repo)
		require.NoError(t, uc.Execute(ctx, "guild1"))
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repo := mockguild.NewMockRepository(t)
		repo.EXPECT().Delete(ctx, "guild1").Return(errors.New("db error"))

		uc := guilduc.NewRemoveAutoRole(repo)
		assert.Error(t, uc.Execute(ctx, "guild1"))
	})
}
