package resttooling

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrors_WithError(t *testing.T) {
	t.Parallel()

	t.Run("ErrorMustFromContext panics when no err in ctx", func(t *testing.T) {
		ctx := context.Background()
		require.Panics(t, func() {
			_ = ErrorMustFromContext(ctx)
		})
	})

	t.Run("it is ok to have nil err in ctx", func(t *testing.T) {
		ctx := context.Background()
		ctx = WithErrorPlaceholder(ctx)
		require.Nil(t, ErrorMustFromContext(ctx))
	})

	t.Run("can put err in ctx", func(t *testing.T) {
		ctx := context.Background()
		ctx = WithErrorPlaceholder(ctx)
		err := errors.New("test_error")
		SetErrorInContext(ctx, err)
		require.Equal(t, err, ErrorMustFromContext(ctx))
	})

	t.Run("errs from ctx are accessible in http handlers up the stack", func(t *testing.T) {
		err := errors.New("test_error")

		h := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			SetErrorInContext(r.Context(), err)
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		*req = *req.WithContext(WithErrorPlaceholder(req.Context()))
		h(nil, req)

		require.Equal(t, err, ErrorMustFromContext(req.Context()))
	})
}
