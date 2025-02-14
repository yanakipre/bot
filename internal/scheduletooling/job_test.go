package scheduletooling

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/testtooling"
)

func TestMultipleExecution(t *testing.T) {
	testtooling.SkipShort(t)

	cfg := logger.DefaultConfig()
	cfg.Format = logger.FormatConsole
	cfg.LogLevel = "ERROR"
	logger.SetNewGlobalLoggerQuietly(cfg)

	for iter := 0; iter < 8; iter++ {
		t.Run(fmt.Sprintf("Iteration%03d", iter), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			n := &atomic.Int64{}
			attempts := &atomic.Int64{}

			t.Cleanup(func() {
				if t.Failed() {
					t.Logf("ran jobs %d times, making %d attempts", n.Load(), attempts.Load())
				}
			})

			ipj := createInProcessJob(t, n, iter)

			wg := &sync.WaitGroup{}
			startJobWorkers(ctx, t, wg, ipj, attempts)

			// kick the scheduler to make sure that at least one of the
			// workers above has run (maybe) before starting below
			runtime.Gosched()

			// check very often that we've only run one job
			checkJobExecution(t, ctx, n)

			cancel()
			wg.Wait()
			verifyJobExecution(t, n, attempts)
		})
	}
}

func createInProcessJob(t *testing.T, n *atomic.Int64, iter int) Job {
	return NewInProcessJob(func(ctx context.Context) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		num := n.Add(1)
		if num != 1 {
			t.Errorf("should only ever get here once, got %d, case=%d:", num, iter)
		}

		t.Cleanup(func() {
			if t.Failed() {
				t.Log("ran job instance", num)
			}
		})

		timer := time.NewTimer(time.Minute)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			if err := ctx.Err(); errors.Is(err, context.DeadlineExceeded) {
				t.Error("should not have timed out")
			}
		case <-timer.C:
			t.Error("should not have reached timeout")
		}
		return nil
	}, Config{UniqueName: t.Name()}, func() (Config, error) { return Config{}, nil }, &metricsCollector{})
}

func startJobWorkers(ctx context.Context, t *testing.T, wg *sync.WaitGroup, ipg Job, attempts *atomic.Int64) {
	// start threads that run jobs
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			timer := time.NewTimer(0)
			defer timer.Stop()
			var count int64
			defer func() {
				attempts.Add(count)
				if count == 0 {
					t.Error("should run at least once")
				}
			}()
			for {
				count++
				select {
				case <-ctx.Done():
					if !t.Failed() && errors.Is(ctx.Err(), context.DeadlineExceeded) {
						t.Error("worker thread should not have hit deadline", count)
					}
					return
				case <-timer.C:
					// try to run
					// the job
					ipg.Execute(ctx)
				}
				// sleep for a jittered amount of
				// time on the order of several milliseconds
				timer.Reset(time.Duration(count+rand.Int63n(20)+1) * time.Millisecond)
			}
		}()
	}
}

func checkJobExecution(t *testing.T, ctx context.Context, n *atomic.Int64) {
	// check very often that we've only run one job
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	const numChecks = 200
	var count int
	for i := 0; i < numChecks; i++ {
		select {
		case <-ticker.C:
			if v := n.Load(); v > 1 {
				t.Error("only one job should run", v)
				break
			}
		case <-ctx.Done():
			t.Error("should not have reached timeout")
			break
		}
		if t.Failed() {
			break
		}
		count++
	}
	if count != numChecks {
		t.Errorf("should have checked %d times, got %d", numChecks, count)
	}
}

func verifyJobExecution(t *testing.T, n *atomic.Int64, attempts *atomic.Int64) {
	nexecs := n.Load()
	if nexecs != 1 {
		t.Error("only one job should run", nexecs)
	}
	nattempts := attempts.Load()
	if nattempts < 100 {
		t.Error("should have attempted to run the test more", nattempts)
	}
}

func TestJobIsolationPanicHandling(t *testing.T) {
	testtooling.SkipShort(t)

	testtooling.SetNewGlobalLoggerQuietly()

	for iter := 0; iter < 4; iter++ {
		t.Run(fmt.Sprintf("Iteration%03d", iter), func(t *testing.T) {
			n := &atomic.Int64{}
			panicCount := &atomic.Int64{}
			job := createPanicJob(t, n, panicCount)

			wg := &sync.WaitGroup{}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// start a bunch of threads that run jobs
			nexecAttempts := &atomic.Int64{}
			startPanicJobWorkers(ctx, t, wg, job, nexecAttempts)

			checkPanicJobExecution(t, ctx, nexecAttempts)

			cancel()
			wg.Wait()

			verifyPanicJobExecution(t, n, panicCount, nexecAttempts)
		})
	}
}

func createPanicJob(t *testing.T, n *atomic.Int64, panicCount *atomic.Int64) Job {
	return NewInProcessJob(func(ctx context.Context) error {
		num := n.Add(1)
		if num%2 == 0 {
			panicCount.Add(1)
			panic("expected panic")
		}

		timer := time.NewTimer(time.Duration(rand.Int63n(100 * int64(time.Millisecond))))
		defer timer.Stop()
		select {
		case <-timer.C:
		case <-ctx.Done():
			if err := ctx.Err(); errors.Is(err, context.DeadlineExceeded) {
				t.Error("should not have timed out")
			}
		}
		return nil
	}, Config{UniqueName: t.Name()}, func() (Config, error) { return Config{}, nil }, &metricsCollector{})
}

func startPanicJobWorkers(ctx context.Context, t *testing.T, wg *sync.WaitGroup, job Job, nexecAttempts *atomic.Int64) {
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			timer := time.NewTimer(0)
			defer timer.Stop()
			var count int64
			defer func() {
				if count == 0 {
					t.Error("should run at least once")
				}
			}()
			for {
				count++
				select {
				case <-timer.C:
					// try to run
					// the job
					nexecAttempts.Add(1)
					job.Execute(ctx)
				case <-ctx.Done():
					if !t.Failed() && errors.Is(ctx.Err(), context.DeadlineExceeded) {
						t.Error("worker thread should not have hit deadline", nexecAttempts.Load())
					}
					return
				}
				// sleep for a jittered amount of
				// time on the order of several milliseconds
				timer.Reset(time.Duration(rand.Int63n(25*int64(time.Millisecond))) + time.Millisecond)
			}
		}()
	}
}

func checkPanicJobExecution(t *testing.T, ctx context.Context, nexecAttempts *atomic.Int64) {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	const numChecks = 200
	for {
		select {
		case <-ctx.Done():
			t.Error("should not have hit timeout")
			return
		case <-ticker.C:
			if nexecAttempts.Load() >= numChecks {
				return
			}
		}
	}
}

func verifyPanicJobExecution(t *testing.T, n *atomic.Int64, panicCount *atomic.Int64, nexecAttempts *atomic.Int64) {
	if n.Load() < 1 {
		t.Error("should have run more than one job", n.Load())
	}
	if nexecAttempts.Load() < 200 {
		t.Errorf("should have called execute at least 200 times, got %d", nexecAttempts.Load())
	}
	if panicCount.Load() < 1 {
		t.Error("should have panic more than once", panicCount.Load(), n.Load())
	}

	computed := n.Load() - panicCount.Load()*2
	if computed > 1 || computed < -1 {
		t.Error("panics should generally be almost half the execs:", map[string]int64{
			"panics":    panicCount.Load(),
			"n":         n.Load(),
			"tolerance": 1,
			"computed":  computed,
		})
	}
}

func TestConcurrentJobExecution(t *testing.T) {
	testtooling.SkipShort(t)

	cfg := logger.DefaultConfig()
	cfg.Format = logger.FormatConsole
	cfg.LogLevel = "ERROR"
	logger.SetNewGlobalLoggerQuietly(cfg)
	concurrentCount := int64(5)
	t.Run("testConcurrentJobExecution", func(t *testing.T) {
		ctx := context.Background()
		n := &atomic.Int64{}

		job := NewConcurrentInProcessJob(func(ctx context.Context) error {
			n.Add(1)
			timer := time.NewTimer(250 * time.Millisecond)
			defer timer.Stop()
			<-timer.C
			return nil
		},
			Config{UniqueName: t.Name()},
			func() (Config, error) { return Config{}, nil },
			&metricsCollector{},
			concurrentCount,
		)

		wg := &sync.WaitGroup{}
		// start goroutines that run jobs
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				job.Execute(ctx)
			}()
		}

		wg.Wait()
		nexecs := n.Load()
		// Assure only 5 jobs ran concurrently
		require.Equal(t, concurrentCount, nexecs)
		ipj := job.(*InProcessJob)
		// Assure that the running count is 0 after all jobs have run
		require.Equal(t, int64(0), ipj.runningCount.Get())
	})
}
