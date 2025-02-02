package sentrytooling

import (
	"context"

	"github.com/getsentry/sentry-go"
)

type sentryCtx string

const sentryCtxKey sentryCtx = "sentry"

// FromContext returns sentry ctx that can be updated after InitCtx has been called
func FromContext(ctx context.Context) sentry.Context {
	return ctx.Value(sentryCtxKey).(sentry.Context)
}

// InitCtx initializes sentry hub and puts Neon-specific data into ctx.
// Ctx can be retrieved and modified later on with FromContext
// keyName should contain "Neon ", e.g. "Neon HTTP" to catch eye of a developer in the UI.
func InitCtx(ctx context.Context, keyName string) context.Context {
	sCtx := sentry.Context{}

	ctx = context.WithValue(ctx, sentryCtxKey, sCtx)

	// `StartSpan()` and `span.Finish()` use `hubFromContext` to get the current hub and
	// store event into it. If hub is missed in the context, then global hub is used and
	// tags are leaking to the hub and will be appeared in all events.
	// `CurrentHub()` returns the global hub, so we need to clone it to use the local one.
	sentryHub := sentry.CurrentHub().Clone()
	sentryScope := sentryHub.Scope()
	if sentryScope == nil {
		sentryScope = sentryHub.PushScope()
	}
	sentryScope.SetContext(keyName, sCtx)
	return sentry.SetHubOnContext(ctx, sentryHub)
}
