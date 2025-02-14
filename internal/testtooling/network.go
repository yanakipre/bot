package testtooling

import (
	"context"

	"github.com/rekby/fixenv"

	"github.com/yanakipre/bot/internal/testtooling/container"
)

func ContainerNetwork(ctx context.Context, e fixenv.Env) *container.Network {
	return fixenv.CacheResult(e, func() (*fixenv.GenericResult[*container.Network], error) {
		net, err := container.NewNetwork(ctx)
		if err != nil {
			e.T().Fatalf("cannot create network: %v", err)
			return nil, err
		}

		cleanup := func() {
			err = net.Close(ctx)
			if err != nil {
				e.T().Fatalf("cannot stop postgres container: %v", err)
			}
		}

		return fixenv.NewGenericResultWithCleanup(net, cleanup), nil
	}, fixenv.CacheOptions{
		Scope:    fixenv.ScopePackage,
		CacheKey: "container-network",
	})
}
