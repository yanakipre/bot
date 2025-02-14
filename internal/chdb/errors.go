package chdb

import "errors"

var (
	ErrConnectionNotInitialized = errors.New("connection not initialized")
	ErrNotAvailableInRoMode     = errors.New("not available in read-only mode")
)
