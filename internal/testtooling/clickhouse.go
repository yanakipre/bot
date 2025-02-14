package testtooling

import (
	"context"
	"fmt"

	"github.com/rekby/fixenv"

	"github.com/yanakipre/bot/internal/chdb"
	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/testtooling/container"
	"github.com/yanakipre/bot/internal/testtooling/datasource"
)

func FixtureClickhouseContainer(ctx context.Context, e fixenv.Env) *container.Clickhouse {
	return fixenv.CacheResult(e, func() (*fixenv.GenericResult[*container.Clickhouse], error) {
		net := ContainerNetwork(ctx, e)

		ch, cleanup, err := RunClickhouseContainer(ctx, net.Name())
		if err != nil {
			e.T().Fatalf("cannot run clickhouse container: %v", err)
			return nil, err
		}

		return fixenv.NewGenericResultWithCleanup(ch, cleanup), nil
	}, fixenv.CacheOptions{
		Scope:    fixenv.ScopePackage,
		CacheKey: "clickhouse-container",
	})
}

func FixtureEmptyClickHouseDB(ctx context.Context, e fixenv.Env, pgDSN string) *chdb.DB {
	return fixenv.CacheResult(e, func() (*fixenv.GenericResult[*chdb.DB], error) {
		chContainer := FixtureClickhouseContainer(ctx, e)

		t := e.T()
		randDbName := DbName("", t.Name())

		ch := datasource.NewClickHouse(chContainer.ExposedAddr().StringWithoutProto())

		if err := ch.CreateDatabase(ctx, randDbName); err != nil {
			e.T().Fatalf("cannot create a new Clickhouse DB: %v", err)
			return nil, err
		}

		chConn, err := ch.Connect(ctx, randDbName)
		if err != nil {
			e.T().Fatalf("cannot connect to Clickhouse: %v", err)
			return nil, err
		}

		if err = ch.InitDBSchema(ctx, pgDSN, randDbName); err != nil {
			e.T().Fatalf("cannot initialize Clickhouse schema: %v", err)
			return nil, err
		}

		cleanup := func() {
			if err = ch.Close(); err != nil {
				e.T().Fatalf("cannot close Clickhouse connection: %v", err)
			}
			if err = ch.DropDatabase(ctx, randDbName); err != nil {
				e.T().Fatalf("cannot drop Clickhouse database: %v", err)
			}
		}

		return fixenv.NewGenericResultWithCleanup(chConn, cleanup), nil
	}, fixenv.CacheOptions{
		Scope:    fixenv.ScopeTest,
		CacheKey: "clickhouse-db",
	})
}

func RunClickhouseContainer(
	ctx context.Context,
	networkName string,
) (*container.Clickhouse, func(), error) {
	chParams := chdb.DefaultConfig()

	ch := container.NewClickhouse(
		datasource.ManagementClickHouseDB,
		chParams.Username,
		chParams.ChcPassword.Unmask(),
		container.WithNetworks(networkName),
	)
	if err := ch.Run(ctx); err != nil {
		return nil, nil, fmt.Errorf("cannot run ch container: %w", err)
	}

	cleanup := func() {
		if err := ch.Stop(ctx); err != nil {
			logger.Error(ctx, fmt.Errorf("cannot stop clickhouse container: %w", err))
		}
	}

	return ch, cleanup, nil
}
