package staticconfig

import (
	"context"
	"errors"
	"fmt"
	"github.com/heetch/confita"
	"github.com/heetch/confita/backend"
)

// Backend specifies custom logic for config loading
type Backend struct{}

func (b Backend) Unmarshal(_ context.Context, to any) error {
	cfg, ok := to.(*Config)
	if !ok {
		return fmt.Errorf("cannot unmarshall to Config: %+v", to)
	}

	cfg.PostgresRW.RDB.DSN.FromEnv("DATABASE_URL")

	return nil
}

// Get is not implemented.
func (b Backend) Get(_ context.Context, _ string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (b Backend) Name() string {
	return "telegramsearch-static"
}

var (
	_ backend.Backend     = &Backend{}
	_ confita.Unmarshaler = &Backend{}
)
