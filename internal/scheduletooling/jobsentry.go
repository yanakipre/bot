package scheduletooling

import (
	"context"

	"github.com/getsentry/sentry-go"

	"github.com/yanakipe/bot/internal/sentrytooling"
)

func (j *InProcessJob) sentryHub(ctx context.Context, traceID string) (context.Context, func()) {
	ctx = sentrytooling.InitCtx(ctx, "Neon Scheduled Job")
	sCtx := sentrytooling.FromContext(ctx)
	sCtx["Trace ID"] = traceID

	sentry.GetHubFromContext(ctx).Scope().SetTransaction(j.Description())

	span := sentry.StartSpan(ctx, "job_execution")
	return ctx, func() {
		span.Finish()
	}
}
