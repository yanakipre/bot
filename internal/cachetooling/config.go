package cachetooling

import "github.com/yanakipre/bot/internal/encodingtooling"

type Config struct {
	// Name is the name of the cache
	Name string `yaml:"-"`
	// InitOnStart is the flag to initialize the cache on service start
	InitOnStart bool `yaml:"init_on_start"`
	// Size is the maximum number of items in the cache
	FullUpdateEnabled bool `yaml:"full_update_enabled"`
	// FullUpdateInterval is the interval of time between full updates.
	FullUpdateInterval encodingtooling.Duration `yaml:"full_update_interval"`
	// FullUpdateBatchSize is the number of items to load in a single batch to fulfil the cache
	// It's not used in console.
	FullUpdateBatchSize uint64 `yaml:"full_update_batch_size"`
	// IncrementalUpdateEnabled is the flag to enable incremental updates
	IncrementalUpdateEnabled bool `yaml:"incremental_update_enabled"`
	// IncrementalUpdateInterval is the interval of time between incremental updates.
	IncrementalUpdateInterval encodingtooling.Duration `yaml:"incremental_update_interval"`
	// IncrementalUpdateBatchSize is the number of items to load in a single batch to fulfil the cache
	IncrementalUpdateBatchSize uint64 `yaml:"incremental_update_batch_size"`
}
