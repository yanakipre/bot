package openapiapp

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/yanakipe/bot/internal/logger"
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
		logger.Error(ctx, "cannot get healthcheck", zap.Error(err))
		return
	}
	marshal, err := json.Marshal(check)
	if err != nil {
		logger.Error(ctx, "cannot marshal healthcheck", zap.Error(err))
		return
	}
	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write(marshal)
	if err != nil {
		logger.Error(ctx, "cannot write response", zap.Error(err))
	}
}
