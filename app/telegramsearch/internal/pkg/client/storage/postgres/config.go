package postgres

import "github.com/yanakipre/bot/internal/rdb"

type Config struct {
	RDB rdb.Config
}

func Default() Config {
	return Config{
		RDB: rdb.DefaultConfig(),
	}
}
