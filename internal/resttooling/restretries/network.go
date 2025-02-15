package restretries

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"syscall"

	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/resttooling/resttimeouts"
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

		if resttimeouts.IsTimeout(err) {
			logger.Info(ctx, "timeout that can be retried", zap.Error(err))
			return err
		}

		if _, ok := idempotentHTTPMethods[req.Method]; ok && adhereToRESTIdempotency {
			logger.Info(
				ctx,
				"network timeout based on HTTP idempotency, retrying",
				zap.String("http_meth", req.Method),
				zap.String("http_path", req.URL.Path),
			)
			return err // safe to retry idempotent HTTP methods.
		}

		slug := getSlug(ctx)
		if _, ok := slugMap[slug]; ok {
			logger.Info(ctx, "network error, retrying", zap.Error(err), zap.String("slug", slug))
			return err // safe to retry
		}

		if isConnectionRefused(err) {
			logger.Info(ctx, "request did not reach the server, safe to retry", zap.Error(err))
			return err // safe to retry
		}
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
	timeoutInterface interface{ Timeout() bool }
)

// isConnectionResetByPeer signals of "read: connection reset by peer" error.
// In other words, client received a RST packet from the server.
//
// That might mean many things, some of which are:
//  1. intermediate proxy broke the connection in flight.
//  2. network errors on switches, etc.
//  3. application acted in a wrong way, completely committing the successful request to it's DB,
//     but could not respond correctly, abruptly breaking TCP connection.
func isConnectionResetByPeer(err error) bool {
	return errors.Is(err, syscall.ECONNRESET)
}

// isConnectionRefused checks if the given error is a "connection refused" error.
//
// This function is useful for determining if a network error is due to a connection
// being refused, which can occur if the server is not accepting connections or if
// there is a network issue preventing the connection from being established.
// We believe that all the servers we're trying to connect to will eventually become online,
// so we consider "connection refused" temporary and safe to retry without thinking of HTTP request semantics.
// Because the request has never reached the server yet.
func isConnectionRefused(err error) bool {
	return errors.Is(err, syscall.ECONNREFUSED)
}

func IsTemporaryNetworkErr(err error) bool {
	if isConnectionResetByPeer(err) {
		return true
	}
	if isConnectionRefused(err) {
		return true
	}
	var timeoutErr timeoutInterface
	if errors.As(err, &timeoutErr) {
		return timeoutErr.Timeout()
	}
	return false
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

// according to RFC-7230
var idempotentHTTPMethods = map[string]struct{}{
	"GET":     {},
	"HEAD":    {},
	"OPTIONS": {},
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

		if resttimeouts.IsTimeout(err) {
			logger.Info(ctx, "timeout from resttimeouts, retrying", zap.Error(err))
			return err
		}

		if isConnectionResetByPeer(err) {
			if _, ok := idempotentHTTPMethods[req.Method]; ok {
				logger.Info(ctx, "connection reset by peer, retrying", zap.Error(err))
				return err // safe to retry idempotent HTTP methods.
			}
			return NewPermanentError(err) // not retrying non-idempotent methods
		}

		if isConnectionRefused(err) {
			if _, ok := idempotentHTTPMethods[req.Method]; ok {
				logger.Info(ctx, "connection refused, retrying", zap.Error(err))
				return err // safe to retry idempotent HTTP methods.
			}
			return NewPermanentError(err) // not retrying non-idempotent methods
		}

		if errors.Is(err, io.ErrUnexpectedEOF) {
			if _, ok := idempotentHTTPMethods[req.Method]; ok {
				logger.Info(ctx, "unexpected EOF, retrying", zap.Error(err))
				return err // safe to retry idempotent HTTP methods.
			}
			return NewPermanentError(err) // not retrying non-idempotent methods
		}

		if isNetworkTimeout(err) {
			if _, ok := idempotentHTTPMethods[req.Method]; ok {
				logger.Info(ctx, "timeout, retrying", zap.Error(err))
				return err // safe to retry idempotent HTTP methods.
			}
			// TODO: #1165 other methods are safe to retry when we have request idempotency.
			return NewPermanentError(err) // not retrying timeouts
		}
		logger.Debug(ctx, "transient network error, retrying", zap.Error(err))

		// any Retry function that returns an error is retried
		// as long as it's not a PermanentError
		return err
	}
}

// UnconditionalNetworkRetry retries all network errors
// Keep in mind, this is NOT a desirable strategy for most APIs.
//
// 1. Because they are often non-idempotent, and blindly retrying them might lead
// to data duplication and internal synchronization conflicts.
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

		logger.Info(ctx, "network error, retrying", zap.Error(err))

		// any Retry function that returns an error is retried
		// as long as it's not a PermanentError
		return err
	}
}
