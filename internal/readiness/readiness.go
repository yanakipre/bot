package readiness

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/kamilsk/retry/v5"
)

type ReadyChecker interface {
	Ready(ctx context.Context) error
}

type state struct {
	checker    ReadyChecker
	isOptional bool
	name       string
}

type Readiness struct {
	checkers   []*state
	strategies retry.How
}

func NewReadiness(
	strategies retry.How,
) *Readiness {
	return &Readiness{
		strategies: strategies,
	}
}

// Add a checker to the readiness probe.
func (r *Readiness) Add(checker ReadyChecker) {
	r.checkers = append(r.checkers, &state{
		checker: checker,
		name:    getType(checker),
	})
}

// AddOptional registers a new optional checker.
// It would run during readiness check, but would not affect readiness state in case of failure.
func (r *Readiness) AddOptional(checker ReadyChecker) {
	r.checkers = append(r.checkers, &state{
		checker:    checker,
		isOptional: true,
		name:       getType(checker),
	})
}

// TryAdd will register ReadyChecker or return false.
func (r *Readiness) TryAdd(candidate any) bool {
	if c, ok := candidate.(ReadyChecker); ok {
		r.Add(c)
		return true
	}
	return false
}

func getType(myvar any) string {
	if t := reflect.TypeOf(myvar); t.Kind() == reflect.Ptr {
		return fmt.Sprintf("%q from %s", t.Elem().Name(), t.Elem().PkgPath())
	} else {
		return fmt.Sprintf("%q from %s", t.Name(), t.PkgPath())
	}
}

// IsReady blocks until all components are ready, or it hits strategy limits.
// returned error is the reason of not being ready.
func (r *Readiness) IsReady(ctx context.Context, log Log) error {
	for _, s := range r.checkers {
		timeout := time.Minute * 2
		if s.isOptional {
			timeout = time.Second * 30
		}
		err := func() error {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			return retry.Do(ctx, func(ctx context.Context) error {
				if err := s.checker.Ready(ctx); err != nil {
					log.NotReady(s.name, s.isOptional, err)
					if s.isOptional {
						return nil // Test other checkers in case of an error.
					}
					return fmt.Errorf("dependency %s could not start: %w", s.name, err)
				}

				log.Ready(s.name)
				return nil
			}, r.strategies...)
		}()
		if err != nil {
			return err
		}
	}
	return nil
}
