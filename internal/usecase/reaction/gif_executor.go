package reaction

import "context"

// GIFExecutor is satisfied by any use case that can fetch a reaction GIF URL.
type GIFExecutor interface {
	Execute(ctx context.Context, reaction string) (string, error)
}
