package worker

import (
	"context"
	"strconv"
	"strings"
)

// WorkloadSplitConfig is for evenly splitting workload among WorkerCount workers
type WorkloadSplitConfig struct {
	WorkerCount uint64 `yaml:"worker_count"           json:"worker_count"`
	// WorkerIndex is being calculated during application startup with GetWorkerIndex
	// It is not expected to be set externally.
	WorkerIndex uint64 `yaml:"worker_index,omitempty" json:"worker_index,omitempty"`
}

// GetWorkerIndex calculates worker index based on hostname
// We rely on kubernetes giving pods in stateful set numeric indexes in hostnames
// like yanakipre-console-api-{0,1, ...}.
// When called from the staticconfig package, no logger is available yet.
func GetWorkerIndex(ctx context.Context, hostname string) (uint64, error) {
	splitParts := strings.Split(hostname, "-")
	if len(splitParts) == 1 {
		// seems we're running locally with a hostname like console.local
		return 0, nil
	}
	parseInt, err := strconv.ParseInt(splitParts[len(splitParts)-1], 10, 64)
	if err != nil {
		// We're running on something like MBP-Anton
		return 0, nil
	}
	return uint64(parseInt), nil
}

func DefaultWorkloadSplit() WorkloadSplitConfig {
	return WorkloadSplitConfig{
		WorkerCount: 1,
		WorkerIndex: 0,
	}
}
