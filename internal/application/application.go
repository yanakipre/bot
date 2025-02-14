package application

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/kamilsk/retry/v5"
	"github.com/kamilsk/retry/v5/backoff"
	"github.com/kamilsk/retry/v5/strategy"
	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/readiness"
	"github.com/yanakipre/bot/internal/scheduletooling"
)

type Application struct {
	readyChecker          *readiness.Readiness
	inProcessJobScheduler *scheduletooling.Scheduler
	shutdownCtx           context.Context
	shutdownMu            sync.Mutex
	components            []Component
	otelShutdown          func(context.Context) error
}

type Component interface {
	StartServer(ctx context.Context)
	ShutdownServer(ctx context.Context)
}

// ReadyCheck registers a service readiness checker.
// NB! The application would not start up if the check fails!
func (a *Application) ReadyCheck(checker readiness.ReadyChecker) {
	a.readyChecker.Add(checker)
}

// OptionalCheck registers an optional service readiness checker.
// If the check fails, it would produce a log message with an error description, but will allow the service to start up.
// It should be used to check non-critical application dependencies during the startup.
func (a *Application) OptionalCheck(checker readiness.ReadyChecker) {
	a.readyChecker.AddOptional(checker)
}

func (a *Application) SetInProcessJobScheduler(scheduler *scheduletooling.Scheduler) {
	a.inProcessJobScheduler = scheduler
}

func (a *Application) AddComponent(c Component) {
	a.components = append(a.components, c)
}

func New(opts ...Option) *Application {
	app := &Application{
		readyChecker: readiness.NewReadiness(retry.How{
			strategy.Backoff(backoff.Incremental(100*time.Millisecond, 100*time.Millisecond)),
		}),
		shutdownCtx: context.Background(),
	}
	for i := range opts {
		opts[i](app)
	}
	return app
}

func (a *Application) IsReady(ctx context.Context) {
	ctx = logger.WithName(ctx, "readiness")
	logger.Info(ctx, "checking critical dependencies")
	if err := a.readyChecker.IsReady(ctx, readiness.Logger(logger.FromContext(ctx))); err != nil {
		logger.Fatal(ctx, "application not ready", zap.Error(err))
	}
	logger.Info(ctx, "critical dependencies are ready")
}

func (a *Application) Start(ctx context.Context) {
	// Sentry setup
	// more details at https://docs.sentry.io/platforms/go/
	if err := sentry.Init(sentry.ClientOptions{
		// Sentry configured from environment variables: SENTRY_DSN, SENTRY_ENVIRONMENT,
		// SENTRY_RELEASE
		Debug:            false,
		AttachStacktrace: true,
	}); err != nil {
		logger.Fatal(ctx, "could not setup sentry", zap.Error(err))
	}

	for i := range a.components {
		go a.components[i].StartServer(ctx)
	}
	if a.inProcessJobScheduler != nil {
		// it must be started before jobs can be added https://github.com/reugn/go-quartz/issues/50
		a.inProcessJobScheduler.Start(logger.WithName(ctx, "background_worker"))
	}
}

// Shutdown makes application stop accepting new traffic.
// It does not close application dependencies.
//
// Those dependencies, that have something long-running code to execute,
// (for example: operation executors, posthog, message bus clients)
// they will impact the application shutdown time a lot and will execute after Shutdown.
func (a *Application) Shutdown(shutdownWait time.Duration) {
	defer sentry.Flush(2 * time.Second)

	a.shutdownMu.Lock()
	defer a.shutdownMu.Unlock()

	shutdownCtx := logger.WithName(a.shutdownCtx, "application")
	logger.Info(shutdownCtx, "application is shutting down",
		zap.Duration("max_wait_time", shutdownWait),
	)
	// We have to put a time limit on shutdown context.
	// If we don't timeout by ourselves, the process manager will kill us with SIGKILL.
	shutdownCtx, cancel := context.WithTimeout(shutdownCtx, shutdownWait)
	defer cancel()

	shutdownStartedAt := time.Now()
	// Shut down all the subsystems in parallel. The order doesn't matter.
	wg := &sync.WaitGroup{}

	for i := range a.components {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.components[i].ShutdownServer(shutdownCtx)
		}()
	}

	if a.inProcessJobScheduler != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.inProcessJobScheduler.Stop()
			a.inProcessJobScheduler.Wait(shutdownCtx)
		}()
	}

	if a.otelShutdown != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := a.otelShutdown(shutdownCtx); err != nil {
				logger.Error(shutdownCtx, fmt.Errorf("could not shut down tracing: %w", err))
			}
		}()
	}

	// wait for everything to close, or the timeout to fire.
	sig := make(chan struct{})
	go func() { defer close(sig); wg.Wait() }()
	select {
	case <-shutdownCtx.Done():
		logger.Info(
			shutdownCtx,
			"did not shutdown application in time",
			zap.Duration("took", time.Since(shutdownStartedAt)),
		)
	case <-sig:
		logger.Info(
			shutdownCtx,
			"clean shutdown for application",
			zap.Duration("took", time.Since(shutdownStartedAt)),
		)
	}
}
