package testtooling

import (
	"testing"
)

// SkipShort skips (sub-) test if -short flag given.
// mb refactor after https://github.com/rekby/fixenv/issues/10?
func SkipShort(t *testing.T) {
	if testing.Short() {
		t.Skip("skip long-running test in short mode")
	}
}
