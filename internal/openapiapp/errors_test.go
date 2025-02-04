package openapiapp

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"

	"github.com/ogen-go/ogen/ogenerrors"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"

	"github.com/yanakipre/bot/internal/semerr"
)

// connectError to test dns errors
type connectError struct {
	msg string
	err error
}

func (e *connectError) Error() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "failed to connect: %s", e.msg)
	if e.err != nil {
		fmt.Fprintf(sb, " (%s)", e.err.Error())
	}
	return sb.String()
}

func (e *connectError) Unwrap() error {
	return e.err
}

func getConnectErr() error {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	d := net.Dialer{}
	_, dnsResErr := d.DialContext(
		ctx,
		"udp",
		"definitelyunknownhost:5333",
	)
	return fmt.Errorf("test: %w", &connectError{
		msg: "test",
		err: dnsResErr,
	})
}

func TestWrapWithSemantic(t *testing.T) {
	type input struct {
		err error
	}
	dnsResErr := getConnectErr()
	tests := []struct {
		name     string
		args     input
		expected semerr.Semantic
	}{
		{
			name: "dns resolution when context is cancelled",
			args: input{
				err: dnsResErr,
			},
			expected: semerr.SemanticCancelled,
		},
		{
			name: "rolled back transaction is cancellation",
			args: input{
				err: &ogenerrors.DecodeParamsError{},
			},
			expected: semerr.SemanticInvalidInput,
		},
		{
			name: "decode params ogen error",
			args: input{
				err: &ogenerrors.DecodeParamsError{},
			},
			expected: semerr.SemanticInvalidInput,
		},
		{
			name: "decode request ogen error",
			args: input{
				err: &ogenerrors.DecodeRequestError{},
			},
			expected: semerr.SemanticInvalidInput,
		},
		{
			name: "security ogen error",
			args: input{
				err: &ogenerrors.SecurityError{},
			},
			expected: semerr.SemanticAuthentication,
		},
		{
			name: "default is 500",
			args: input{
				err: xerrors.New("test"),
			},
			expected: semerr.SemanticInternal,
		},
		{
			name: "deadline exceeded handled",
			args: input{
				err: context.Canceled,
			},
			expected: semerr.SemanticCancelled,
		},
		{
			name: "semantic handled",
			args: input{
				err: semerr.NotFound("test"),
			},
			expected: semerr.SemanticNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := WrapWithSemantic(tt.args.err)
			require.Equal(t, tt.expected, actual.Semantic)
		})
	}
}
