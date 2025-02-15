package restretries

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/clouderr"
	"github.com/yanakipre/bot/internal/logger"
)

// RetryByStatusCodeFunc signals that request should be retried
// if response status code satisfies condition
func RetryByStatusCodeFunc(shouldRetry func(responseCode int) bool) RetryByResponse {
	return func(ctx context.Context, response *http.Response) error {
		if shouldRetry(response.StatusCode) {
			logger.Info(ctx, "retrying by status code",
				zap.Int("code", response.StatusCode))
			return clouderr.WithFields("response status code is retryable", zap.Int("status", response.StatusCode))
		}
		return nil
	}
}

// RetryByStatusCode signals that request should be retried if response finished with code from
// retryableCodes.
func RetryByStatusCode(retryableCodes ...int) RetryByResponse {
	return RetryByStatusCodeFunc(func(responseCode int) bool {
		for _, code := range retryableCodes {
			if code == responseCode {
				return true
			}
		}
		return false
	})
}

// RepeatRetriableStatusCodes will repeat requests that are safe to retry from console standpoint.
func RepeatRetriableStatusCodes() RetryByResponse {
	return RetryByStatusCode(
		http.StatusServiceUnavailable,
		http.StatusTooManyRequests,
		http.StatusBadGateway,
		http.StatusLocked,
	)
}
