package testtooling

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/yanakipe/bot/internal/semerr"
)

// CompareEqual compares, respecting Equal method of types.
//
// This allows, for example, comparison of same timestamps,
// presented by time.Time from different time.Location to be equal.
func CompareEqual(t *testing.T, exp, actual any, opts ...cmp.Option) {
	require.Equal(t, cmp.Equal(exp, actual, opts...), true, cmp.Diff(exp, actual, opts...))
}

// IsSemantic asserts that error passes IsSemantic
var IsSemantic = func(t require.TestingT, err error) *semerr.Error {
	require.Error(t, err)
	serr := semerr.AsSemanticError(err)
	require.NotNil(t, serr)
	return serr
}

func MustValue[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
