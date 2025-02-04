package postgres

import (
	"context"
	"time"

	"github.com/yanakipre/bot/internal/encodingtooling"
	"github.com/yanakipre/bot/internal/rdb"
)

type Storage struct {
	cfg Config
	// now generates current time.
	now func() time.Time
	db  *rdb.DB
}

func New(
	cfg Config,
) *Storage {
	db := rdb.New(cfg.RDB)
	return &Storage{
		cfg: cfg,
		now: time.Now,
		db:  db,
	}
}

// Ready implements check for readinesschecker
func (s *Storage) Ready(ctx context.Context) error {
	if err := s.db.Ready(ctx); err != nil {
		return err
	}
	s.db.MapperFunc(encodingtooling.CamelToSnake)
	return nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}
