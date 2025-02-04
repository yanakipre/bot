package resttooling

import (
	"context"

	"github.com/yanakipre/bot/internal/resttooling/roundtripper"
)

type uriSlug string

var slugURIKey uriSlug = "uri-slug"

// URISlugToContext persists slug in the context to be further used when producing metrics.
func URISlugToContext(ctx context.Context, slug string) context.Context {
	return context.WithValue(ctx, slugURIKey, slug)
}

// URISlugFromContext extracts slug that was put by URISlugToContext.
// If none, returns empty string.
func URISlugFromContext(ctx context.Context) (uri string) {
	if val := ctx.Value(slugURIKey); val != nil {
		return val.(string)
	}
	return ""
}

// MetricReportCfg preconfigures the way to report metrics.
// It's intended to be used exclusively with MetricFromContext.
type MetricReportCfg struct {
	ClientName string
	StandID    string
}

// MetricFromContext is a wrapper to gather metrics data.
// It's intended to be used along with URISlugToContext.
//
// If you need something more specialised,
// be inspired by MetricFromContext and write your own.
func MetricFromContext(cfg MetricReportCfg) func(ctx context.Context) (e roundtripper.MetricEvent) {
	return func(ctx context.Context) (e roundtripper.MetricEvent) {
		return roundtripper.MetricEvent{
			StandID: cfg.StandID,
			Service: cfg.ClientName,
			Uri:     URISlugFromContext(ctx),
		}
	}
}
