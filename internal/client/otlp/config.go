package otlp

// Configuration options for OpenTelemetry OTLP exporter.
//
// This is used in two places in the static config file:
//
//  1. In the otlp_exporter field. That is used to configure the exporter of
//     yanakipre-console itself.
//  2. In operations.compute_otlp_exporter. That is passed through to the 'compute_ctl'
//     binary, to configure tracing in the compute node.
//
// Usually you want the traces to end up in the same place, and you want to set
// both to the same value. But if you use a collector proxy in one but not the other,
// for example, then you might need to set them differently.
type OtlpExporterCfg struct {
	Name    string
	Enabled bool
	// default is taken from OTEL_EXPORTER_OTLP_ENDPOINT env variable, if empty
	Endpoint string
}

func DefaultOtlpExporterCfg() OtlpExporterCfg {
	return OtlpExporterCfg{
		Name:     "unnamed-service",
		Enabled:  true,
		Endpoint: "http://jaeger:4318",
	}
}
