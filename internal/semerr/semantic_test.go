package semerr

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestIsSemanticError(t *testing.T) {
	t.Parallel()
	inputs := []struct {
		Err      error
		Semantic Semantic
		Is       bool
	}{
		{
			Err:      NotImplemented("not implemented"),
			Semantic: SemanticNotImplemented,
			Is:       true,
		},
		{
			Err:      Authentication("authentication"),
			Semantic: SemanticNotImplemented,
		},
		{
			Err:      errors.New("foo"),
			Semantic: SemanticNotImplemented,
		},
		{
			Err:      fmt.Errorf("bar: %w", NotImplemented("not implemented")),
			Semantic: SemanticNotImplemented,
			Is:       true,
		},
		{
			Err:      WrapWithNotImplemented(errors.New("bar"), "not implemented"),
			Semantic: SemanticNotImplemented,
			Is:       true,
		},
		{
			Err: WrapWithUnavailable(fmt.Errorf("bar: %w", NotImplemented("not implemented")),
				"unavailable"),
			Semantic: SemanticUnavailable,
			Is:       true,
		},
		{
			Err: WrapWithUnavailable(fmt.Errorf("bar: %w", NotImplemented("not implemented")),
				"unavailable"),
			Semantic: SemanticNotImplemented,
		},
		{
			Err: WrapWithUnavailable(fmt.Errorf("bar: %w", NotImplemented("not implemented")),
				"unavailable"),
			Semantic: SemanticAuthentication,
		},
	}

	for _, input := range inputs {
		t.Run(fmt.Sprintf("%s/%d/%t", input.Err, input.Semantic, input.Is), func(t *testing.T) {
			assert.Equal(t, input.Is, IsSemanticError(input.Err, input.Semantic))
		})
	}
}

func TestAsSemanticError(t *testing.T) {
	inputs := []struct {
		Err error
		As  bool
	}{
		{
			Err: NotImplemented("foo"),
			As:  true,
		},
		{
			Err: Authentication("foo"),
			As:  true,
		},
		{
			Err: errors.New("foo"),
		},
		{
			Err: fmt.Errorf("bar: %w", NotImplemented("foo")),
			As:  true,
		},
	}

	for _, input := range inputs {
		t.Run(fmt.Sprintf("%+v", input.Err), func(t *testing.T) {
			assert.Equal(t, input.As, AsSemanticError(input.Err) != nil)
		})
	}
}

func TestMultiWrapIs(t *testing.T) {
	err := WrapWithNotFound(
		WrapWithNotImplemented(
			WrapWithUnavailable(
				io.EOF,
				"unavailable",
			),
			"not implemented",
		),
		"not found",
	)
	assert.True(t, errors.Is(err, io.EOF))
	assert.False(t, errors.Is(err, errors.New("random")))

	assert.True(t, IsNotFound(err))
	assert.True(t, IsSemanticError(err, SemanticNotFound))

	assert.False(t, IsNotImplemented(err))
	assert.False(t, IsSemanticError(err, SemanticNotImplemented))

	assert.False(t, IsUnavailable(err))
	assert.False(t, IsSemanticError(err, SemanticUnavailable))

	assert.False(t, IsAuthentication(err))
	assert.False(t, IsSemanticError(err, SemanticAuthentication))

	assert.False(t, IsForbidden(err))
	assert.False(t, IsSemanticError(err, SemanticForbidden))
}

func TestWrapErrorfFormatting(t *testing.T) {
	err := wrapErrorf(SemanticUnavailable, io.EOF, "foo: %s", "bar")
	assert.ErrorIs(t, err, io.EOF)
	assert.Equal(t, "foo: bar: EOF", err.Error())
}

func TestWrapWithFields(t *testing.T) {
	err := WithFields("origin", zap.String("k1", "v1"))
	err = fmt.Errorf("wrapped1: %w", err)
	err = WrapWithFields(err)
	err = WrapWithFields(err, zap.String("k1", "ignored1"), zap.String("k2", "v2"))
	err = fmt.Errorf(
		"wrapped2: %w",
		WrapWithFields(err, zap.String("k1", "ignored2"), zap.String("k3", "v3"), zap.String("k4", "v4")),
	)
	err = WrapWithInternal(err, "internal error")
	err = WrapWithFields(WrapWithFields(err, zap.String("k1", "ignored3")), zap.String("k1", "ignored4"))
	err = errors.Join(err, WithFields("joined", zap.String("k5", "v5")))

	require.True(t, IsSemanticError(err, SemanticInternal))
	require.Equal(t, "internal error: wrapped2: wrapped1: origin\njoined", err.Error())

	require.Len(t, UnwrapFields(err), 5)
	require.Contains(
		t,
		UnwrapFields(err),
		zap.String("k1", "v1"),
		zap.String("k2", "v2"),
		zap.String("k3", "v3"),
		zap.String("k4", "v4"),
		zap.String("k5", "v5"),
	)
}
