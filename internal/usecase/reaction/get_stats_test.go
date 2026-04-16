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

func TestGetStats_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("returns stats from repository", func(t *testing.T) {
		repo := mockreaction.NewMockStatsRepository(t)
		expected := map[string]int64{"hug": 10, "pat": 3}
		repo.EXPECT().GetAll(ctx).Return(expected, nil)

		uc := reactionuc.NewGetStats(repo)
		stats, err := uc.Execute(ctx)
		require.NoError(t, err)
		assert.Equal(t, expected, stats)
	})

	t.Run("returns empty map when no stats recorded", func(t *testing.T) {
		repo := mockreaction.NewMockStatsRepository(t)
		repo.EXPECT().GetAll(ctx).Return(map[string]int64{}, nil)

		uc := reactionuc.NewGetStats(repo)
		stats, err := uc.Execute(ctx)
		require.NoError(t, err)
		assert.Empty(t, stats)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repo := mockreaction.NewMockStatsRepository(t)
		repoErr := errors.New("db error")
		repo.EXPECT().GetAll(ctx).Return(nil, repoErr)

		uc := reactionuc.NewGetStats(repo)
		stats, err := uc.Execute(ctx)
		assert.ErrorIs(t, err, repoErr)
		assert.Nil(t, stats)
	})
}
