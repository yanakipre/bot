package httpserver

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"go.uber.org/zap"

	"github.com/yanakipe/bot/internal/logger"
)

// StartServer starts the weserver and blocks.
// One has to pass correct context with predefined rules for cancellation.
func StartServer(ctx context.Context, s *http.Server, appName string) {
	s.BaseContext = func(net.Listener) context.Context { return ctx }
	logger.Info(ctx, fmt.Sprintf("starting %s on: %s", appName, s.Addr))
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal(ctx, fmt.Sprintf("%s error", appName), zap.Error(err))
	}
}
