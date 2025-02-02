package unittooling

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestCast(t *testing.T) {
	t.Parallel()
	require.Equal(t, "1.8500", FormatFloat4(*BytesToGiB(1986422374)))
	require.Equal(t, uint64(1986422374), *BytesFromGiB(1.85))
	require.Equal(t, uint64(2*1024*1024), *BytesFromMiB(2))
	require.Equal(t, "1.2000", FormatFloat4(*SecondsToHours(4320)))
	require.Equal(t, uint64(4320), *SecondsFromHours(1.2))
	require.Equal(t, 1.2, *CentsToDollars(120))
	require.Equal(t, uint64(120), DollarsToCents(decimal.NewFromFloat(1.2)))
	require.Equal(t, uint64(126), DollarsToCents(decimal.NewFromFloat(1.256)))
	require.Equal(t, "53.3198", FormatFloat4(53.31984))
	val, err := DollarsStrToCents("53.31984")
	require.NoError(t, err)
	require.Equal(t, uint64(5332), val)
}
