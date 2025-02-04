package testtooling

import (
	"context"

	"github.com/rekby/fixenv"

	"github.com/yanakipre/bot/internal/testtooling/container"
)

func ContainerNetwork(ctx context.Context, e fixenv.Env) *container.Network {
	cacheKey := "container-network"
	return e.CacheWithCleanup(cacheKey, &fixenv.FixtureOptions{
		Scope: fixenv.ScopePackage,
	}, func() (any, fixenv.FixtureCleanupFunc, error) {
		net, err := container.NewNetwork(ctx)
		if err != nil {
			e.T().Fatalf("cannot create network: %v", err)
			return nil, nil, err
		}

		cleanup := func() {
			err = net.Close(ctx)
			if err != nil {
				e.T().Fatalf("cannot stop postgres container: %v", err)
			}
		}

		return net, cleanup, nil
	}).(*container.Network)
}
