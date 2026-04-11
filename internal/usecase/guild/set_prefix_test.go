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

func TestSetPrefix_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("saves guild with prefix", func(t *testing.T) {
		repo := mockguild.NewMockRepository(t)
		repo.EXPECT().Save(ctx, &domain.Guild{ID: "guild1", Prefix: "!"}).Return(nil)

		uc := guilduc.NewSetPrefix(repo)
		require.NoError(t, uc.Execute(ctx, "guild1", "!"))
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repo := mockguild.NewMockRepository(t)
		repo.EXPECT().Save(ctx, &domain.Guild{ID: "guild1", Prefix: "!"}).
			Return(errors.New("db error"))

		uc := guilduc.NewSetPrefix(repo)
		assert.Error(t, uc.Execute(ctx, "guild1", "!"))
	})
}
