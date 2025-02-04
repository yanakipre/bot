package sentrytooling

import (
	"context"
	"errors"
	"syscall"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"

	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/semerr"
)

// these semantic errors will be sent to Sentry.
var noSkipSentryForSemerrs = []semerr.Semantic{
	semerr.SemanticUnknown,
	semerr.SemanticInternal,
}

var errSkipSentry = errors.New("this skips sentry")

// SkipSentry signals that this error should not get into sentry
func SkipSentry(err error) error {
	return errors.Join(err, errSkipSentry)
}

// shouldSkipSentry
//
// Sentry is to see business logic problems.
// So we restrict the scope of errors that is to be send to Sentry.
func shouldSkipSentry(err error) bool {
	if errors.Is(err, errSkipSentry) {
		return true
	}

	serr := semerr.AsSemanticError(err)
	if serr == nil {
		return false
	}

	switch {
	case errors.Is(err, context.Canceled):
		// Do not report events "canceled by client"
		return true

	case errors.Is(err, context.DeadlineExceeded):
		// Do not report events of timeouts waiting for some other applications.
		// They happen and we expect them to affect SLO.
		// If they don't affect SLO they do not matter at all.
		return true

	case errors.Is(err, syscall.ECONNRESET):
		fallthrough
	case errors.Is(err, syscall.EPIPE):
		// Don't report broken pipes or connection resets - the other party in a connection disconnected.
		// There's not much we can do and it usually appears when the client disconnects before reading the whole response.
		return true

	default:
		return !slices.Contains(noSkipSentryForSemerrs, serr.Semantic)
	}
}

// Report is like Sentry, but accepts any errors.
func Report(ctx context.Context, err error) {
	if shouldSkipSentry(err) {
		return
	}

	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.CaptureException(err)
	} else {
		logger.Error(ctx, "could not get sentry hub from context", zap.Error(err))
	}
}

// Sentry is a helper function that reports semantic errors to Sentry.
func Sentry(ctx context.Context, err *semerr.Error) {
	Report(ctx, err)
}
