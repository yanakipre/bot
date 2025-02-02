package slicetooling

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClosest(t *testing.T) {
	tests := []struct {
		name       string
		batchSizes []int
		num        int
		want       int
	}{
		{
			name:       "positive batch sizes",
			batchSizes: []int{1, 5, 10, 20},
			num:        7,
			want:       5,
		},
		{
			name:       "negative batch sizes",
			batchSizes: []int{-20, -10, -5, -1},
			num:        -7,
			want:       -5,
		},
		{
			name:       "mixed batch sizes",
			batchSizes: []int{-10, -5, 0, 5, 10},
			num:        3,
			want:       5,
		},
		{
			name:       "num is in batch sizes",
			batchSizes: []int{1, 5, 10, 20},
			num:        10,
			want:       10,
		},
		{
			name:       "empty batch sizes",
			batchSizes: []int{},
			num:        7,
			want:       7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Closest(tt.batchSizes, tt.num)
			require.Equal(t, tt.want, got)
		})
	}
}
