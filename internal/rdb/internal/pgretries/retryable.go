package pgretries

import (
	"context"
	"errors"
	"net"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// isRetryable treats errors as following:
// 1. context cancellation and deadline exceeded as non-retryable
// 2. DNS/Network errors as retryable
// 3. Postgres errors as retryable if they are connection failures
func isRetryable(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var dnserr *net.DNSError
	if errors.As(err, &dnserr) {
		// all but context cancellation errors are retryable
		return !strings.Contains(dnserr.Err, "operation was canceled")
	}
	neterr := &net.OpError{}
	if errors.As(err, &neterr) {
		return true
	}
	pgerr := &pgconn.PgError{}
	if errors.As(err, &pgerr) {
		switch pgerr.Code {
		case pgerrcode.ConnectionFailure,
			pgerrcode.ConnectionException,
			pgerrcode.AdminShutdown,
			pgerrcode.OperatorIntervention,
			pgerrcode.CrashShutdown,
			// https://www.metisdata.io/knowledgebase/errors/postgresql-53300
			pgerrcode.TooManyConnections,
			// https://www.metisdata.io/knowledgebase/errors/postgresql-57p03
			pgerrcode.CannotConnectNow:
			return true
		default:
			return false
		}
	}
	return errors.Is(err, &pgconn.ConnectError{})
}
