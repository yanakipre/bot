package testtooling

import (
	"context"
	"fmt"

	"github.com/rekby/fixenv"
	"go.uber.org/zap"

	"github.com/yanakipe/bot/internal/chdb"
	"github.com/yanakipe/bot/internal/logger"
	"github.com/yanakipe/bot/internal/testtooling/container"
	"github.com/yanakipe/bot/internal/testtooling/datasource"
)

func FixtureClickhouseContainer(ctx context.Context, e fixenv.Env) *container.Clickhouse {
	cacheKey := "clickhouse-container"
	return e.CacheWithCleanup(cacheKey, &fixenv.FixtureOptions{
		Scope: fixenv.ScopePackage,
	}, func() (any, fixenv.FixtureCleanupFunc, error) {
		net := ContainerNetwork(ctx, e)

		ch, cleanup, err := RunClickhouseContainer(ctx, net.Name())
		if err != nil {
			e.T().Fatalf("cannot run clickhouse container: %v", err)
			return nil, nil, err
		}

		return ch, cleanup, nil
	}).(*container.Clickhouse)
}

func FixtureEmptyClickHouseDB(ctx context.Context, e fixenv.Env, pgDSN string) *chdb.DB {
	cacheKey := "clickhouse-db"
	return e.CacheWithCleanup(cacheKey, &fixenv.FixtureOptions{
		Scope: fixenv.ScopeTest,
	}, func() (any, fixenv.FixtureCleanupFunc, error) {
		chContainer := FixtureClickhouseContainer(ctx, e)

		t := e.T()
		randName := DbName("", t.Name())

		ch := datasource.NewClickHouse(chContainer.ExposedAddr().StringWithoutProto())
		if err := ch.CreateDatabase(ctx, randName); err != nil {
			e.T().Fatalf("cannot create a new Clickhouse DB: %v", err)
			return nil, nil, err
		}

		chConn, err := ch.Connect(ctx, randName)
		if err != nil {
			e.T().Fatalf("cannot connect to Clickhouse: %v", err)
			return nil, nil, err
		}

		if err = ch.InitDBSchema(ctx, pgDSN); err != nil {
			e.T().Fatalf("cannot initialize Clickhouse schema: %v", err)
			return nil, nil, err
		}

		cleanup := func() {
			if err = ch.Close(); err != nil {
				e.T().Fatalf("cannot close Clickhouse connection: %v", err)
			}
			if err = ch.DropDatabase(ctx, randName); err != nil {
				e.T().Fatalf("cannot drop Clickhouse database: %v", err)
			}
		}

		return chConn, cleanup, nil
	}).(*chdb.DB)
}

func RunClickhouseContainer(
	ctx context.Context,
	networkName string,
) (*container.Clickhouse, func(), error) {
	chParams := chdb.DefaultConfig()

	zk := container.NewZookeeper(container.WithNetworks(networkName))
	if err := zk.Run(ctx); err != nil {
		return nil, nil, fmt.Errorf("cannot initialize zookeeper container: %w", err)
	}

	ch := container.NewClickhouse(
		datasource.ManagementClickHouseDB,
		chParams.Username,
		chParams.Password.Unmask(),
		container.WithNetworks(networkName),
	)
	if err := ch.Run(ctx); err != nil {
		return nil, nil, fmt.Errorf("cannot run zookeeper container: %w", err)
	}

	cleanup := func() {
		if err := ch.Stop(ctx); err != nil {
			logger.Error(ctx, "cannot stop clickhouse container", zap.Error(err))
		}
		if err := zk.Stop(ctx); err != nil {
			logger.Error(ctx, "cannot stop zookeeper container", zap.Error(err))
		}
	}

	return ch, cleanup, nil
}
