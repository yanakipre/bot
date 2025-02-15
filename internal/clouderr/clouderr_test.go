package clouderr_test

import (
	"errors"
	"fmt"
	"github.com/yanakipre/bot/internal/clouderr"
	"github.com/yanakipre/bot/internal/semerr"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestWrapWithFields(t *testing.T) {
	{
		err := clouderr.WithFields("origin", zap.String("k1", "v1"))
		err = fmt.Errorf("wrapped1: %w", err)
		err = clouderr.WrapWithFields(err)
		err = clouderr.WrapWithFields(err, zap.String("k1", "ignored1"), zap.String("k2", "v2"))
		err = fmt.Errorf(
			"wrapped2: %w",
			clouderr.WrapWithFields(err, zap.String("k1", "ignored2"), zap.String("k3", "v3"), zap.String("k4", "v4")),
		)
		err = semerr.WrapWithInternal(err, "internal error")
		err = clouderr.WrapWithFields(
			clouderr.WrapWithFields(err, zap.String("k1", "ignored3")),
			zap.String("k1", "ignored4"),
		)
		err = errors.Join(err, clouderr.WithFields("joined", zap.String("k5", "v5")))

		require.True(t, semerr.IsSemanticError(err, semerr.SemanticInternal))
		require.Equal(t, "internal error: wrapped2: wrapped1: origin\njoined", err.Error())

		require.Len(t, clouderr.UnwrapFields(err), 5)
		require.ElementsMatch(
			t,
			clouderr.UnwrapFields(err),
			[]zap.Field{
				zap.String("k1", "v1"),
				zap.String("k2", "v2"),
				zap.String("k3", "v3"),
				zap.String("k4", "v4"),
				zap.String("k5", "v5"),
			},
		)
	}

	{
		err := error(semerr.Internal("origin", zap.String("k1", "v1")))
		err = clouderr.WrapWithFields(err, zap.String("k1", "ignored1"), zap.String("k2", "v2"))

		require.True(t, semerr.IsSemanticError(err, semerr.SemanticInternal))
		require.Equal(t, "origin", err.Error())

		require.Len(t, clouderr.UnwrapFields(err), 2)
		require.ElementsMatch(
			t,
			clouderr.UnwrapFields(err),
			[]zap.Field{
				zap.String("k1", "v1"),
				zap.String("k2", "v2"),
			},
		)
	}
}

func TestFieldToString(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2024-08-06T14:01:59.398423Z")
	require.NoError(t, err)

	for _, tc := range []struct {
		field    zap.Field
		expected string
	}{
		{zap.String("key", "world"), `key:"world"`},
		{zap.ByteString("key", []byte("world")), `key:"world"`},
		{zap.Int("key", 404), `key:"404"`},
		{zap.Bool("key", true), `key:"true"`},
		{zap.Time("key", testTime), `key:"2024-08-06 14:01:59.398423 +0000 UTC"`},
		{zap.Duration("key", 10*time.Second), `key:"10s"`},
		{zap.Float64("key", 1.23), `key:"1.23"`},
		{zap.Binary("key", []byte{1, 2, 3}), `key:"010203"`},
	} {
		assert.Equal(t, tc.expected, clouderr.FieldToString(tc.field), "unexpected output for %#v", tc.field)
	}
}

func TestNilError(t *testing.T) {
	require.Panics(t, func() {
		_ = clouderr.WrapWithFields(nil, zap.String("k1", "v1"))
	})
	require.Panics(t, func() {
		_ = clouderr.WrapWithFields(nil)
	})
}
