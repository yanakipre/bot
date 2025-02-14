package sentrytooling

import (
	"context"
	"errors"
	"fmt"
	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/semerr"
	"github.com/yanakipre/bot/internal/status"
	"github.com/yanakipre/bot/internal/status/codes"
	"slices"
	"syscall"

	"github.com/getsentry/sentry-go"
	"github.com/ogen-go/ogen/ogenerrors"
)

// These types of errors will be sent to Sentry
var (
	noSkipSentryForSemerrs = []semerr.Semantic{
		semerr.SemanticInternal,
		semerr.SemanticUnknown,
	}
	noSkipSentryForStatusCodes = []codes.Code{
		codes.Internal,
		codes.Unknown,
	}
)

var errSkipSentry = errors.New("this skips sentry")

// SkipSentry signals that this error should not get into sentry
func SkipSentry(err error) error {
	return errors.Join(err, errSkipSentry)
}

// shouldSkipSentry
//
// Sentry is to see business logic problems.
// So we restrict the scope of errors that are to be sent to Sentry.
func shouldSkipSentry(err error) bool {
	if errors.Is(err, errSkipSentry) {
		return true
	}

	var oerr ogenerrors.Error
	if ok := errors.As(err, &oerr); ok {
		// Skip errors that are due to request decoding, e.g., missing fields in the request.
		return true
	}

	switch {
	case errors.Is(err, context.Canceled):
		// Do not report events "canceled by client"
		return true

	case errors.Is(err, context.DeadlineExceeded):
		// Do not report events of timeouts waiting for some other applications.
		// They happen and we expect them to affect SLO.
		// If they don't affect SLO, they do not matter at all.
		return true

	case errors.Is(err, syscall.ECONNRESET):
		fallthrough
	case errors.Is(err, syscall.EPIPE):
		// Don't report broken pipes or connection resets - the other party in a connection disconnected.
		// There's little we can do,
		// and it usually appears when the client disconnects before reading the whole response.
		return true
	}

	var st *status.Error
	if ok := errors.As(err, &st); ok {
		return !slices.Contains(noSkipSentryForStatusCodes, st.Status().Code())
	}

	serr := semerr.AsSemanticError(err)
	if serr == nil {
		return false
	}

	return !slices.Contains(noSkipSentryForSemerrs, serr.Semantic)
}

// Report is a helper function that reports semantic errors to Sentry.
func Report(ctx context.Context, err error) {
	if shouldSkipSentry(err) {
		return
	}

	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.CaptureException(err)
	} else {
		logger.Error(
			ctx,
			fmt.Errorf("could not get sentry hub from context: %w", err),
		)
	}
}
