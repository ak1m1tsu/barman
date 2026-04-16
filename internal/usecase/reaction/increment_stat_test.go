package reaction_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	reactionuc "github.com/ak1m1tsu/barman/internal/usecase/reaction"
	mockreaction "github.com/ak1m1tsu/barman/mocks/reaction"
)

func TestIncrementStat_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("increments successfully", func(t *testing.T) {
		repo := mockreaction.NewMockStatsRepository(t)
		repo.EXPECT().Increment(ctx, "hug").Return(nil)

		uc := reactionuc.NewIncrementStat(repo)
		err := uc.Execute(ctx, "hug")
		require.NoError(t, err)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repo := mockreaction.NewMockStatsRepository(t)
		repoErr := errors.New("db error")
		repo.EXPECT().Increment(ctx, "hug").Return(repoErr)

		uc := reactionuc.NewIncrementStat(repo)
		err := uc.Execute(ctx, "hug")
		assert.ErrorIs(t, err, repoErr)
	})
}
