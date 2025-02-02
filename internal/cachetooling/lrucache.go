package cachetooling

import lru "github.com/hashicorp/golang-lru/v2"

func NewLRU[K comparable, V any](cfg LRUConfig) (*lru.Cache[K, V], error) {
	return lru.New[K, V](cfg.Size)
}

type LRUConfig struct {
	Size int `yaml:"size"`
}
