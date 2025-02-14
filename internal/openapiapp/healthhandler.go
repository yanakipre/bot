package openapiapp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yanakipre/bot/internal/logger"
)

type HealthCheckFunc func(ctx context.Context) (any, error)

// HealthHandler handles /healthz endpoint.
type healthCheckHandler struct {
	getHealth func(ctx context.Context) (any, error)
}

func (h healthCheckHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	check, err := h.getHealth(ctx)
	if err != nil {
		logger.Error(ctx, fmt.Errorf("cannot get healthcheck: %w", err))
		return
	}
	marshal, err := json.Marshal(check)
	if err != nil {
		logger.Error(ctx, fmt.Errorf("cannot marshal healthcheck: %w", err))
		return
	}
	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write(marshal)
	if err != nil {
		logger.Error(ctx, fmt.Errorf("cannot write response: %w", err))
	}
}
