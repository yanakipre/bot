package slicetooling

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Difference(t *testing.T) {
	type left struct {
		ID     string
		field1 int
	}

	type right struct {
		ID     string
		field2 []string
	}

	tests := []struct {
		name      string
		leftOnes  []left
		rightOnes []right
		wantLeft  []left
		wantRight []right
	}{
		{
			name: "empty",
		},
		{
			name: "left empty",
			rightOnes: []right{
				{ID: "1", field2: []string{"a"}},
				{ID: "2", field2: []string{"b"}},
			},
			wantRight: []right{
				{ID: "1", field2: []string{"a"}},
				{ID: "2", field2: []string{"b"}},
			},
		},
		{
			name: "right empty",
			leftOnes: []left{
				{ID: "1", field1: 1},
				{ID: "2", field1: 2},
			},
			wantLeft: []left{
				{ID: "1", field1: 1},
				{ID: "2", field1: 2},
			},
		},
		{
			name: "no intersection",
			leftOnes: []left{
				{ID: "1"},
			},
			rightOnes: []right{
				{ID: "2", field2: []string{"b"}},
			},
			wantLeft: []left{
				{ID: "1"},
			},
			wantRight: []right{
				{ID: "2", field2: []string{"b"}},
			},
		},
		{
			name: "with intersection",
			leftOnes: []left{
				{ID: "1"},
				{ID: "3"},
			},
			rightOnes: []right{
				{ID: "2", field2: []string{"b"}},
				{ID: "3"},
			},
			wantLeft: []left{
				{ID: "1"},
			},
			wantRight: []right{
				{ID: "2", field2: []string{"b"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id1 := func(i left) string { return i.ID }
			id2 := func(i right) string { return i.ID }
			left, right := Difference(tt.leftOnes, tt.rightOnes, id1, id2)
			assert.ElementsMatch(t, tt.wantLeft, left)
			assert.ElementsMatch(t, tt.wantRight, right)
		})
	}
}
