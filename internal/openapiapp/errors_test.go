package openapiapp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-faster/jx"
	"github.com/ogen-go/ogen/ogenerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"

	"github.com/yanakipre/bot/internal/codeerr"
	"github.com/yanakipre/bot/internal/resttooling"
	"github.com/yanakipre/bot/internal/semerr"
	"github.com/yanakipre/bot/internal/status/codes"
	"github.com/yanakipre/bot/internal/testtooling"
)

type TestErrorStatusCode struct {
	StatusCode int
	Response   TestError
}

// GetStatusCode returns the value of StatusCode.
func (s *TestErrorStatusCode) GetStatusCode() int {
	return s.StatusCode
}

// GetResponse returns the value of Response.
func (s *TestErrorStatusCode) GetResponse() TestError {
	return s.Response
}

type TestError struct {
	Error string `json:"error"`
}

func (s *TestError) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

func (s *TestError) Encode(e *jx.Encoder) {
	e.ObjStart()
	s.encodeFields(e)
	e.ObjEnd()
}

// encodeFields encodes fields.
func (s *TestError) encodeFields(e *jx.Encoder) {
	e.FieldStart("error")
	e.Str(s.Error)
}

func Test_ErrorHandler(t *testing.T) {
	h := ErrorHandler(func(ctx context.Context, err error) *TestErrorStatusCode {
		return &TestErrorStatusCode{
			StatusCode: 418,
			Response: TestError{
				Error: "not an espresso machine",
			},
		}
	})

	appErr := errors.New("internal error")

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/coffee", nil)
	h(context.Background(), w, r, appErr)

	assert.Equal(t, 418, w.Code)
	assert.Equal(t, "application/json", w.Result().Header.Get("Content-Type"))
	assert.Equal(t, `{"error":"not an espresso machine"}`, w.Body.String())
}

func Test_PresentError(t *testing.T) {
	tests := []struct {
		name             string
		err              error
		expectStatusCode int
		expectErrorCode  codeerr.ErrorCode
		expectMessage    string
		expectErr        error
	}{
		{
			name:             "raw error",
			err:              errors.New("test_error"),
			expectStatusCode: 500,
			expectMessage:    "unknown internal server error",
		},
		{
			name:             "sematic error",
			err:              semerr.NotFound("test_error"),
			expectStatusCode: 404,
			expectMessage:    "test_error",
		},
		{
			name:             "code err with semantic err",
			err:              codeerr.Wrap(codeerr.ProjectsLimitExceeded, semerr.NotFound("test_error")),
			expectStatusCode: 404,
			expectErrorCode:  codeerr.ProjectsLimitExceeded,
			expectMessage:    "test_error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testtooling.SetNewGlobalLoggerQuietly()

			ctx := resttooling.WithErrorPlaceholder(context.Background())

			statusCode, errorCode, message := PresentError(ctx, tt.err)
			require.Equal(t, tt.expectStatusCode, statusCode)
			require.Equal(t, tt.expectErrorCode, errorCode)
			require.Equal(t, tt.expectMessage, message)
			var sem *semerr.Error
			require.ErrorAs(t, resttooling.ErrorMustFromContext(ctx), &sem)
			require.Equal(t, wrapWithSemantic(tt.err).MessageWithFields(), sem.MessageWithFields())
		})
	}
}

func Test_PresentStatus(t *testing.T) {
	tests := []struct {
		name            string
		err             error
		expectCode      codes.Code
		expectErrorCode codeerr.ErrorCode
		expectMessage   string
	}{
		{
			name:          "raw error",
			err:           errors.New("test_error"),
			expectCode:    codes.Unknown,
			expectMessage: "unknown error",
		},
		{
			name:          "sematic error",
			err:           semerr.NotFound("test_error"),
			expectCode:    codes.NotFound,
			expectMessage: "test_error",
		},
		{
			name:            "code err with semantic err",
			err:             codeerr.Wrap(codeerr.ProjectsLimitExceeded, semerr.NotFound("test_error")),
			expectCode:      codes.NotFound,
			expectErrorCode: codeerr.ProjectsLimitExceeded,
			expectMessage:   "test_error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testtooling.SetNewGlobalLoggerQuietly()

			ctx := resttooling.WithErrorPlaceholder(context.Background())

			st, errorCode := PresentStatus(ctx, tt.err)
			require.Equal(t, tt.expectCode, st.Code())
			require.Equal(t, tt.expectErrorCode, errorCode)
			require.Equal(t, tt.expectMessage, st.Message())
			require.Equal(t, st.Error(), resttooling.ErrorMustFromContext(ctx))
		})
	}
}

func Test_payloadFromSemantic(t *testing.T) {
	type args struct {
		err *semerr.Error
	}
	tests := []struct {
		name   string
		args   args
		expect string
	}{
		{
			name: "hides internal from users",
			args: args{
				err: semerr.WrapWithNotFound(errors.New("users not see"), "test not found"),
			},
			expect: "test not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := payloadFromSemantic(tt.args.err)
			require.Equal(t, tt.expect, got)
		})
	}
}

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
	tests := []struct {
		name     string
		args     input
		expected semerr.Semantic
	}{
		{
			name: "dns resolution when context is canceled",
			args: input{
				err: getConnectErr(),
			},
			expected: semerr.SemanticCanceled,
		},
		{
			name: "dns resolution when context is canceled (explicit value)",
			args: input{
				err: &net.DNSError{Err: "dial udp 172.20.0.10:53: operation was canceled"},
			},
			expected: semerr.SemanticCanceled,
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
			expected: semerr.SemanticCanceled,
		},
		{
			name: "semantic handled",
			args: input{
				err: semerr.NotFound("test"),
			},
			expected: semerr.SemanticNotFound,
		},
		{
			name: "semantic survives codeerr",
			args: input{
				err: codeerr.Wrap(codeerr.ProjectsLimitExceeded, semerr.NotFound("test")),
			},
			expected: semerr.SemanticNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := wrapWithSemantic(tt.args.err)
			require.Equal(t, tt.expected, actual.Semantic)
		})
	}
}
