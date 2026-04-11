package reaction

import "context"

// GIFFetcher fetches a GIF URL for a named reaction.
type GIFFetcher interface {
	Fetch(ctx context.Context, reaction string) (string, error)
}

// FetchGIFUseCase fetches a reaction GIF URL.
type FetchGIFUseCase struct {
	fetcher GIFFetcher
}

func NewFetchGIF(f GIFFetcher) *FetchGIFUseCase {
	return &FetchGIFUseCase{fetcher: f}
}

func (uc *FetchGIFUseCase) Execute(ctx context.Context, reaction string) (string, error) {
	return uc.fetcher.Fetch(ctx, reaction)
}
