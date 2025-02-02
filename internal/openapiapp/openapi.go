package openapiapp

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"go.uber.org/zap"

	"github.com/yanakipe/bot/internal/logger"
)

type App struct {
	cfg Config
	s   *http.Server

	// This handles the requests. When you create a new App by calling New(), the Mux
	// is filled in with a rule for /healthz, and for cfg.baseURL. You may add
	// additional rules before calling StartServer()
	Mux *http.ServeMux
}

// StartServer binds server to a port and waits for it to finish.
// It is expected to be run in separate goroutine.
func (a *App) StartServer(ctx context.Context) {
	// we want values from context but not the cncellation
	// this should be long living context for graceful shutdown to work
	baseCtx := context.WithoutCancel(ctx)
	a.s.BaseContext = func(net.Listener) context.Context { return baseCtx }
	logger.Info(ctx, fmt.Sprintf("starting %q on: %s", a.cfg.Name, a.s.Addr))
	if err := a.s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal(ctx, fmt.Sprintf("%q error", a.cfg.Name), zap.Error(err))
	}
}

func (a *App) ShutdownServer(ctx context.Context) {
	if err := a.s.Shutdown(ctx); err != nil && err != ctx.Err() {
		logger.Error(ctx, fmt.Sprintf("could not shutdown %s server", a.cfg.Name), zap.Error(err))
	}
}

func New(
	cfg Config,
	handler http.Handler,
	mv []Middleware,
	getHealth HealthCheckFunc,
) *App {
	handlerWithMiddlewares := Wrap(
		handler,
		// first middleware in the list is executed last
		mv...,
	)

	// All server responds from BaseURL, except for healthz endpoint.
	mux := http.NewServeMux()

	if cfg.BaseURL != "" {
		mux.Handle(cfg.BaseURL+"/", http.StripPrefix(cfg.BaseURL, handlerWithMiddlewares))
	} else {
		mux.Handle("/", handlerWithMiddlewares)
	}
	mux.Handle("/healthz", healthCheckHandler{getHealth: getHealth})

	return &App{
		s: &http.Server{
			Addr:              cfg.Addr,
			Handler:           mux,
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		},
		cfg: cfg,
		Mux: mux,
	}
}
