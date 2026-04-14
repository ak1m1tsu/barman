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

func TestFetchGIFWithFallback_Execute(t *testing.T) {
	ctx := context.Background()
	primaryURL := "https://nekos.best/hug/1.gif"
	fallbackURL := "https://otakugifs.xyz/hug/1.gif"

	t.Run("returns primary url when primary succeeds", func(t *testing.T) {
		primary := mockreaction.NewMockGIFFetcher(t)
		fallback := mockreaction.NewMockGIFFetcher(t)
		primary.EXPECT().Fetch(ctx, "hug").Return(primaryURL, nil)
		fallback.EXPECT().Fetch(ctx, "hug").Return(fallbackURL, nil)

		uc := reactionuc.NewFetchGIFWithFallback(primary, fallback)
		url, err := uc.Execute(ctx, "hug")
		require.NoError(t, err)
		assert.Equal(t, primaryURL, url)
	})

	t.Run("returns fallback url when primary fails", func(t *testing.T) {
		primary := mockreaction.NewMockGIFFetcher(t)
		fallback := mockreaction.NewMockGIFFetcher(t)
		primary.EXPECT().Fetch(ctx, "lick").Return("", errors.New("404 not found"))
		fallback.EXPECT().Fetch(ctx, "lick").Return(fallbackURL, nil)

		uc := reactionuc.NewFetchGIFWithFallback(primary, fallback)
		url, err := uc.Execute(ctx, "lick")
		require.NoError(t, err)
		assert.Equal(t, fallbackURL, url)
	})

	t.Run("returns error when both fail", func(t *testing.T) {
		primary := mockreaction.NewMockGIFFetcher(t)
		fallback := mockreaction.NewMockGIFFetcher(t)
		primary.EXPECT().Fetch(ctx, "unknown").Return("", errors.New("primary error"))
		fallback.EXPECT().Fetch(ctx, "unknown").Return("", errors.New("fallback error"))

		uc := reactionuc.NewFetchGIFWithFallback(primary, fallback)
		_, err := uc.Execute(ctx, "unknown")
		assert.Error(t, err)
	})
}
