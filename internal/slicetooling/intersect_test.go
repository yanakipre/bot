package slicetooling

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Intersect(t *testing.T) {
	type value struct {
		ID    string
		value int
	}

	tests := []struct {
		name      string
		leftOnes  []value
		rightOnes []value
		want      []value
	}{
		{
			name: "empty",
		},
		{
			name: "left empty",
			rightOnes: []value{
				{ID: "1", value: 1},
				{ID: "2", value: 1},
			},
			want: nil,
		},
		{
			name: "right empty",
			leftOnes: []value{
				{ID: "1", value: 1},
				{ID: "2", value: 1},
			},
			want: nil,
		},
		{
			name: "no intersection",
			leftOnes: []value{
				{ID: "1"},
			},
			rightOnes: []value{
				{ID: "2", value: 1},
			},
			want: nil,
		},
		{
			name: "with intersection should take right",
			leftOnes: []value{
				{ID: "1"},
				{ID: "3", value: 1},
				{ID: "4", value: 1},
			},
			rightOnes: []value{
				{ID: "2"},
				{ID: "3", value: 2},
				{ID: "4", value: 1},
			},
			want: []value{
				{ID: "3", value: 2},
				{ID: "4", value: 1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idFunc := func(i value) string { return i.ID }
			values := IntersectTakeRight(tt.leftOnes, tt.rightOnes, idFunc)
			assert.ElementsMatch(t, tt.want, values)
		})
	}
}
