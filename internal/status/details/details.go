// Package details provides structures for common error details used by the Status package.
package details

import (
	"time"

	"github.com/yanakipre/bot/internal/status/details/reason"
)

// ErrorDetails is a set of potential error details that can be included to provide more context to the caller.
type ErrorDetails struct {
	ErrorInfo         *ErrorInfo
	RetryInfo         *RetryInfo
	UserFacingMessage *UserFacingMessage
}

// ErrorInfo provides a machine-readable proximate cause of the error with additional, optional metadata.
type ErrorInfo struct {
	Reason   reason.Reason
	Metadata map[string]any
}

// RetryInfo provides the minimum delay before a retry should be attempted.
type RetryInfo struct {
	RetryDelay time.Duration
}

// UserFacingMessage provides a human-readable message that can be shown to the end-user.
type UserFacingMessage struct {
	Message string
}
