package sentrytooling

import (
	"context"
	"errors"
	"net"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yanakipe/bot/internal/openapiapp"
	"github.com/yanakipe/bot/internal/semerr"
)

func Test_shouldSkipSentry(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		skipsSentry bool
	}{
		{
			name:        "simple error goes to sentry",
			err:         errors.New("some err"),
			skipsSentry: false,
		},
		{
			name: "SkipSentry makes semantic Internal errors avoid sentry",
			err: semerr.WrapWithInternal(
				SkipSentry(errors.New("some err")),
				"test-auth-err",
			),
			skipsSentry: true,
		},
		{
			name:        "SkipSentry makes error avoid sentry",
			err:         SkipSentry(errors.New("some err")),
			skipsSentry: true,
		},
		{
			name:        "skips context.Canceled errors wrapped in semantic errors",
			err:         semerr.WrapWithInternal(context.Canceled, "canceled"),
			skipsSentry: true,
		},
		{
			name:        "skips context.DeadlineExceeded errors wrapped in semantic errors",
			err:         semerr.WrapWithInternal(context.DeadlineExceeded, "timeout"),
			skipsSentry: true,
		},
		{
			name:        "skips context.DeadlineExceeded errors wrapped in semantic errors",
			err:         semerr.WrapWithInternal(context.DeadlineExceeded, "timeout"),
			skipsSentry: true,
		},
		{
			name:        "skips connection reset errors wrapped in semantic errors",
			err:         semerr.WrapWithInternal(&net.OpError{Err: syscall.ECONNRESET}, "write"),
			skipsSentry: true,
		},
		{
			name: "skips EPIPE errors wrapped in semantic errors",
			// EPIPE would normally be wrapped in net.OpError
			err:         semerr.WrapWithInternal(&net.OpError{Err: syscall.EPIPE}, "write"),
			skipsSentry: true,
		},
		{
			name:        "skips other semantic errors",
			err:         semerr.InvalidInput("foo"),
			skipsSentry: true,
		},
		{
			name:        "doesn't skip internal errors otherwise",
			err:         semerr.Internal("foo"),
			skipsSentry: false,
		},
		{
			name: "skips DNS errors",
			err: openapiapp.WrapWithSemantic(
				&net.DNSError{Err: "dial udp 172.20.0.10:53: operation was canceled"},
			),
			skipsSentry: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.skipsSentry, shouldSkipSentry(tt.err))
		})
	}
}
