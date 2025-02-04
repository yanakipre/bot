package testtooling

import (
	"context"
	"fmt"

	"github.com/orlangure/gnomock"
	gnomockredis "github.com/orlangure/gnomock/preset/redis"

	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/redis"
	"github.com/yanakipre/bot/internal/secret"
)

func StartRedis(ctx context.Context) (*gnomock.Container, *redis.Redis) {
	p := gnomockredis.Preset(
		gnomockredis.WithVersion("6.0.9"),
	)
	logger.Info(ctx, "redis starting")
	container, err := gnomock.Start(p)
	if err != nil {
		panic(fmt.Sprintf("could not start redis project %v", err))
	}
	rdb, err := redis.New(redis.Config{
		AuthType: redis.AuthPlain,
		URL: secret.NewString(
			fmt.Sprintf("redis://%s:%d", container.Host, container.DefaultPort()),
		),
		ClientName: "yanakipre-test",
	}, ctx)
	if err != nil {
		panic(err)
	}
	logger.Info(ctx, fmt.Sprintf("redis started at %q", rdb.Options().Addr))
	return container, rdb
}
