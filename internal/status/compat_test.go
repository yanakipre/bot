package status

import (
	"context"
	"fmt"
	"net"
	"syscall"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/yanakipre/bot/internal/semerr"
	"github.com/yanakipre/bot/internal/status/codes"
)

func TestFromError(t *testing.T) {
	tests := map[string]struct {
		err   error
		check func(t *testing.T, in error, out *Status)
	}{
		"connection reset is Unavailable": {
			// when some dependency is restarting/crashlooping it might drop all the connections in flight,
			// it's OK to treat those kind of errors as Unavailable
			err: fmt.Errorf("wrapped error: %w", syscall.ECONNRESET),
			check: func(t *testing.T, in error, out *Status) {
				require.Equal(t, out.Code(), codes.Unavailable)
			},
		},
		"connection refused is Unavailable": {
			// when some dependency is restarting, it's OK to receive connection refused for a short period of time
			// until the dependency binds to the port
			err: fmt.Errorf("wrapped error: %w", syscall.ECONNREFUSED),
			check: func(t *testing.T, in error, out *Status) {
				require.Equal(t, out.Code(), codes.Unavailable)
			},
		},
		"postgres CannotConnectNow is Unavailable": {
			err: &pgconn.PgError{Code: pgerrcode.CannotConnectNow},
			check: func(t *testing.T, in error, out *Status) {
				require.Equal(t, out.Code(), codes.Unavailable)
			},
		},
		"postgres TooManyConnections is Unavailable": {
			err: &pgconn.PgError{Code: pgerrcode.TooManyConnections},
			check: func(t *testing.T, in error, out *Status) {
				require.Equal(t, out.Code(), codes.Unavailable)
			},
		},
		"supports semerr": {
			err: semerr.NotFound("not found"),
			check: func(t *testing.T, in error, out *Status) {
				require.Equal(t, out.Code(), codes.NotFound)
				require.True(t, semerr.IsSemanticError(out.Error(), semerr.SemanticNotFound))
			},
		},
		"doesn't unwrap errors into known types": {
			err: fmt.Errorf("oh no: %w", semerr.InvalidInput("invalid input")),
			check: func(t *testing.T, in error, out *Status) {
				// Previously, status would only keep the inner error - now we expect it to keep the outer one,
				// of an unknown type, as well.
				require.ErrorIs(t, out.Error(), in)
			},
		},
		"extracts Statuses from wrapper errors": {
			err: fmt.Errorf("oh no: %w", New(codes.NotFound, "not found").Error()),
			check: func(t *testing.T, in error, out *Status) {
				// We actually do want to extract the status here
				require.NotErrorIs(t, out.Error(), in)
				require.Equal(t, out.Code(), codes.NotFound)
			},
		},
		"supports context.Canceled": {
			err: context.Canceled,
			check: func(t *testing.T, in error, out *Status) {
				require.Equal(t, out.Code(), codes.Canceled)
			},
		},
		"supports context.DeadlineExceeded": {
			err: context.DeadlineExceeded,
			check: func(t *testing.T, in error, out *Status) {
				require.Equal(t, out.Code(), codes.DeadlineExceeded)
			},
		},
		"doesn't override context errors with other errors": {
			err: semerr.WrapWithInternal(context.Canceled, "internal error"),
			check: func(t *testing.T, in error, out *Status) {
				require.Equal(t, out.Code(), codes.Canceled)
			},
		},
		"doesn't override dns errors with other errors": {
			err: semerr.WrapWithInternal(lo.ToPtr(
				net.DNSError{
					Err: "operation was canceled",
				},
			), "internal error"),
			check: func(t *testing.T, in error, out *Status) {
				require.Equal(t, out.Code(), codes.Canceled)
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			st := FromError(tt.err)
			tt.check(t, tt.err, st)
		})
	}
}
