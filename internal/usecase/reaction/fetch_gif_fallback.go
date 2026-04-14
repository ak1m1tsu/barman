package reaction

import "context"

// FetchGIFWithFallbackUseCase fetches a reaction GIF from the primary source and,
// if that fails, from the fallback source. Both requests are fired in parallel.
type FetchGIFWithFallbackUseCase struct {
	primary  GIFFetcher
	fallback GIFFetcher
}

func NewFetchGIFWithFallback(primary, fallback GIFFetcher) *FetchGIFWithFallbackUseCase {
	return &FetchGIFWithFallbackUseCase{primary: primary, fallback: fallback}
}

func (uc *FetchGIFWithFallbackUseCase) Execute(ctx context.Context, reaction string) (string, error) {
	type res struct {
		url string
		err error
	}

	primaryCh := make(chan res, 1)
	fallbackCh := make(chan res, 1)

	go func() {
		url, err := uc.primary.Fetch(ctx, reaction)
		primaryCh <- res{url, err}
	}()
	go func() {
		url, err := uc.fallback.Fetch(ctx, reaction)
		fallbackCh <- res{url, err}
	}()

	if r := <-primaryCh; r.err == nil {
		return r.url, nil
	}
	r := <-fallbackCh
	return r.url, r.err
}
