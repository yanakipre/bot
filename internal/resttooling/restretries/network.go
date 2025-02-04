package restretries

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/logger"
)

// URISlugBasedNetworkRetry is a network error handling strategy for REST.
// Pass all the slugs you want to be retryable.
// Pass the way to retrieve slug as a first argument.
func URISlugBasedNetworkRetry(
	slugs []string,
	getSlug func(ctx context.Context) string,
	adhereToRESTIdempotency bool,
) RetryByNetwork {
	slugMap := make(map[string]bool, len(slugs))
	for _, s := range slugs {
		slugMap[s] = true
	}
	return func(err error, req *http.Request) error {
		if err == nil {
			return nil
		}

		ctx := req.Context()

		if err := checkContext(ctx, err); err != nil {
			return NewPermanentError(err) // not retrying context cancellation
		}

		if err := ignoreTemporaryNetworkErr(err); err != nil {
			return NewPermanentError(err) // only retry temporary errors
		}

		if adhereToRESTIdempotency && (req.Method == "GET" || req.Method == "HEAD") {
			logger.Debug(
				ctx,
				"network timeout based on HTTP idempotency, retrying",
				zap.String("method", req.Method),
			)
			return err // safe to retry idempotent HTTP methods.
		}

		slug := getSlug(ctx)
		if _, ok := slugMap[slug]; ok {
			logger.Debug(ctx, "network error, retrying", zap.Error(err))
			return err // safe to retry
		}
		logger.Warn(ctx, "network error, unsafe to retry")
		return NewPermanentError(err) // not retrying timeouts
	}
}

func checkContext(ctx context.Context, err error) error {
	if ctx.Err() != nil && !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		if err != nil {
			return fmt.Errorf("context canceled with underlying error: %w", err)
		}
		return ctx.Err()
	}

	if errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}

type (
	timeoutInterface   interface{ Timeout() bool }
	temporaryInterface interface{ Temporary() bool }
)

func IsTemporaryNetworkErr(err error) bool {
	var timeoutErr timeoutInterface
	if errors.As(err, &timeoutErr) {
		return timeoutErr.Timeout()
	}
	var temporaryErr temporaryInterface
	if errors.As(err, &temporaryErr) {
		return temporaryErr.Temporary()
	}
	return false
}

// check if an error reports being temporary.
// ignoreTemporaryNetworkErr only returns err when it's not temporary.
func ignoreTemporaryNetworkErr(err error) error {
	if !IsTemporaryNetworkErr(err) {
		return err
	}

	return nil
}

// isNetworkTimeout knows about net/http errors
// and returns true if timeout inside of net/http happened.
func isNetworkTimeout(err error) bool {
	// TODO: consider dropping this retry case. canceled requests
	// were probably canceled for a reason (and were probably
	// filtered out already by checking the context.)
	if strings.Contains(err.Error(), "request canceled") {
		// This is straight from net/http unexported error. This is a client request cancellation
		// error.
		return true
	}

	// if a request has timed out, it's probably not worth
	// retrying but may be safe for idempotent requests.
	var timeoutErr timeoutInterface
	return errors.As(err, &timeoutErr) && timeoutErr.Timeout()
}

// StraightforwardNetworkRetry is a common sense network error handling strategy for REST.
// Does not retry on non-idempotent calls.
func StraightforwardNetworkRetry() RetryByNetwork {
	return func(err error, req *http.Request) error {
		if err == nil {
			return nil
		}

		ctx := req.Context()

		if err := checkContext(ctx, err); err != nil {
			return NewPermanentError(err) // not retrying context cancellation
		}

		if err := ignoreTemporaryNetworkErr(err); err != nil {
			return NewPermanentError(err) // only retry temporary errors
		}

		if isNetworkTimeout(err) {
			if req.Method == "GET" || req.Method == "HEAD" {
				logger.Debug(ctx, "timeout, retrying")
				return err // safe to retry idempotent HTTP methods.
			}
			// TODO: #1165 other methods are safe to retry when we have request idempotency.
			logger.Warn(ctx, "network timeout, unsafe to retry")
			return NewPermanentError(err) // not retrying timeouts
		}
		logger.Debug(ctx, "network error, retrying", zap.Error(err))

		// any Retry function that returns an error is retried
		// as long as it's not a PermanentError
		return err
	}
}

// UnconditionalNetworkRetry retries all network errors
// Keep in mind, this is NOT a desirable strategy for most APIs.
//
// 1. Because they are often non-idempotent, and blindly retrying them might lead
// to data duplication and internal synchronisation conflicts.
// And even if you are implementing ONLY idempotent methods from the API,
// there is NO guarantee someone after you knows about that and will not implement
// non-idempotent method to the client.
//
// 2. If domain is misconfigured, it will continue retrying,
// because it's unconditional.
//
// So, much safer StraightforwardNetworkRetry should be considered
// if you're not sure what you're doing.
func UnconditionalNetworkRetry() RetryByNetwork {
	return func(err error, req *http.Request) error {
		if err == nil {
			return nil
		}

		ctx := req.Context()
		if err := checkContext(ctx, err); err != nil {
			return NewPermanentError(err) // not retrying context cancellation
		}

		logger.Debug(ctx, "network error, retrying", zap.Error(err))

		// any Retry function that returns an error is retried
		// as long as it's not a PermanentError
		return err
	}
}
