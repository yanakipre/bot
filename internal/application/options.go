package application

import (
	"context"

	"go.uber.org/zap"

	"github.com/yanakipe/bot/internal/client/otlp"
	"github.com/yanakipe/bot/internal/logger"
)

type Option func(a *Application)

func WithOpenTelemetry(ctx context.Context, cfg otlp.OtlpExporterCfg) Option {
	return func(a *Application) {
		// Set up opentelemetry tracing
		otelShutdown, err := otlp.InstallExportPipeline(ctx, cfg)
		if err != nil {
			logger.Fatal(ctx, "could initialize tracing", zap.Error(err))
		}
		a.otelShutdown = otelShutdown
	}
}
