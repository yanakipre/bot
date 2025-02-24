package buses

import "context"

type Client interface {
	GetNearest(ctx context.Context, dot Dot) ([]Bus, error)
}
