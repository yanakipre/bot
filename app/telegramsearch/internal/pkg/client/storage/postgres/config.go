package postgres

import "github.com/yanakipe/bot/internal/rdb"

type Config struct {
	RDB rdb.Config
}

func Default() Config {
	return Config{
		RDB: rdb.DefaultConfig(),
	}
}
