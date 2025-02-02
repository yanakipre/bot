package restretries

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/yanakipe/bot/internal/logger"
)

// RetryByStatusCodeFunc signals that request should be retried
// if response status code satisfies condition
func RetryByStatusCodeFunc(shouldRetry func(responseCode int) bool) RetryByResponse {
	return func(ctx context.Context, response *http.Response) error {
		if shouldRetry(response.StatusCode) {
			logger.Info(ctx, "retrying by status code",
				zap.Int("code", response.StatusCode))
			return fmt.Errorf("status code %d is retryable", response.StatusCode)
		}
		return nil
	}
}

// RetryByStatusCodeFuncWithBodyLogging signals that request should be retried
// if response status code satisfies condition
func RetryByStatusCodeFuncWithBodyLogging(shouldRetry func(responseCode int) bool) RetryByResponse {
	return func(ctx context.Context, response *http.Response) error {
		if shouldRetry(response.StatusCode) {
			var bodyStr string
			if response.Body != nil {
				defer func() {
					err := response.Body.Close()
					if err != nil {
						logger.Warn(ctx, "failed to close response body", zap.Error(err))
					}
				}()
				// body can only be read once, but we are going for retry anyway
				body, err := io.ReadAll(response.Body)
				if err != nil {
					logger.Warn(ctx, "failed to read response body", zap.Error(err))
				}
				bodyStr = string(body)
			}

			logger.Info(ctx, "retrying by status code",
				zap.Int("code", response.StatusCode),
				zap.String("body", bodyStr))
			return fmt.Errorf("status code %d is retryable", response.StatusCode)
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
	)
}
