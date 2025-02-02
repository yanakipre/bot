package unittooling

import (
	"strconv"

	"github.com/shopspring/decimal"
)

const (
	secondsInHour = 60 * 60
	bytesInGiB    = 1024 * 1024 * 1024
	bytesInMiB    = 1024 * 1024
	centsInDollar = 100
)

func BytesToGiB(b uint64) *float64 {
	result := float64(b) / bytesInGiB
	return &result
}

func BytesFromGiB(gib float64) *uint64 {
	result := uint64(gib * bytesInGiB)
	return &result
}

func BytesFromMiB(mib int64) *uint64 {
	result := uint64(mib * bytesInMiB)
	return &result
}

func SecondsToHours(s uint64) *float64 {
	result := float64(s) / secondsInHour
	return &result
}

func SecondsFromHours(hours float64) *uint64 {
	result := uint64(hours * secondsInHour)
	return &result
}

func CentsToDollars(c uint64) *float64 {
	result := float64(c) / centsInDollar
	return &result
}

func DollarsToCents(d decimal.Decimal) uint64 {
	return d.Mul(decimal.NewFromInt(centsInDollar)).Round(0).BigInt().Uint64()
}

func DollarsStrToCents(d string) (uint64, error) {
	result, err := decimal.NewFromString(d)
	if err != nil {
		return 0, err
	}
	return DollarsToCents(result), nil
}

func FormatFloat4(f float64) string {
	return strconv.FormatFloat(f, 'f', 4, 64)
}
