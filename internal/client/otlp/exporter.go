package otlp

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"

	"github.com/yanakipre/bot/internal/logger"
)

// Set up OpenTelemetry tracing and HTTP exporter
//
// Use the config file to configure the endpoint. The exporter will pick
// up any other configuration from global defaults and environment variables.
// (https://opentelemetry.io/docs/concepts/sdk-configuration/otlp-exporter-configuration/)
// But if you need to change any other settings, it's better to add it to the
// console config file than use the env variables.
func InstallExportPipeline(
	ctx context.Context,
	cfg OtlpExporterCfg,
) (func(context.Context) error, error) {
	if !cfg.Enabled {
		logger.Info(ctx, "OpenTelemetry HTTP exporter is disabled")
		return func(context.Context) error { return nil }, nil
	}

	logger.Info(
		ctx,
		fmt.Sprintf(
			"initializing OpenTelemetry HTTP exporter with endpoint %s",
			string(cfg.Endpoint),
		),
	)

	opts := []otlptracehttp.Option{}

	// Split the endpoint URL into parts
	if cfg.Endpoint != "" {
		endpointUrl, err := url.Parse(cfg.Endpoint)
		if err != nil {
			return nil, fmt.Errorf(
				"could not parse OTLP endpoint setting \"%s\": %w",
				cfg.Endpoint,
				err,
			)
		}
		opts = append(opts, otlptracehttp.WithEndpoint(endpointUrl.Host))
		if endpointUrl.Path != "" {
			opts = append(opts, otlptracehttp.WithURLPath(endpointUrl.Path))
		}
		if endpointUrl.Scheme != "https" {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
	}

	client := otlptracehttp.NewClient(opts...)
	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP trace exporter: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(cfg.Name),
				semconv.ServiceVersionKey.String(os.Getenv("APP_VERSION")),
			),
		),
	)

	otel.SetTracerProvider(tracerProvider)

	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tracerProvider.Shutdown, nil
}
