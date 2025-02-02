package wrapper

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Hook common interface.
type Hook interface {
	Before(context.Context) context.Context
	After(context.Context, error)
}

type connIDHook struct {
	connID string
	hook   Hook
}

// ConnIDFromContext get connection id from context.
func ConnIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(connIDKeyType(0)).(string); ok {
		return v
	}
	return ""
}

type connIDKeyType int

func (h connIDHook) Before(ctx context.Context) context.Context {
	return h.hook.Before(context.WithValue(ctx, connIDKeyType(0), h.connID))
}

func (h connIDHook) After(ctx context.Context, err error) {
	h.hook.After(ctx, err)
}

var globalRand = rand.New(&lockedSource{src: rand.NewSource(time.Now().UnixNano())}) //nolint:gosec

// lockedSource allows a random number generator to be used by multiple goroutines concurrently.
// The code is very similar to math/rand.lockedSource, which is unfortunately not exposed.
type lockedSource struct {
	mut sync.Mutex
	src rand.Source
}

func (r *lockedSource) Int63() (n int64) {
	r.mut.Lock()
	n = r.src.Int63()
	r.mut.Unlock()
	return
}

// Seed implements Seed() of Source
func (r *lockedSource) Seed(seed int64) {
	r.mut.Lock()
	r.src.Seed(seed)
	r.mut.Unlock()
}

func genConnID() (uuid string) {
	return fmt.Sprintf("%x", globalRand.Int63())
}
