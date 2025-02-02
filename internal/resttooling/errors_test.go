package resttooling

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrors_WithError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	require.Panics(t, func() {
		_ = ErrorMustFromContext(ctx)
	})

	ctx = WithError(ctx, nil)
	require.Nil(t, ErrorMustFromContext(ctx))

	err := errors.New("test_error")
	ctx = WithError(ctx, err)
	require.Equal(t, err, ErrorMustFromContext(ctx))
}
