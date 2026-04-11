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

func TestFetchGIF_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("returns gif url on success", func(t *testing.T) {
		fetcher := mockreaction.NewMockGIFFetcher(t)
		fetcher.EXPECT().Fetch(ctx, "hug").Return("https://nekos.best/api/v2/hug/test.gif", nil)

		uc := reactionuc.NewFetchGIF(fetcher)
		url, err := uc.Execute(ctx, "hug")
		require.NoError(t, err)
		assert.Equal(t, "https://nekos.best/api/v2/hug/test.gif", url)
	})

	t.Run("propagates fetcher error", func(t *testing.T) {
		fetcher := mockreaction.NewMockGIFFetcher(t)
		fetcher.EXPECT().Fetch(ctx, "hug").Return("", errors.New("service unavailable"))

		uc := reactionuc.NewFetchGIF(fetcher)
		_, err := uc.Execute(ctx, "hug")
		assert.Error(t, err)
	})
}
