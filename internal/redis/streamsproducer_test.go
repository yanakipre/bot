package redis

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRedisMessageID(t *testing.T) {
	type testCase struct {
		name     string
		input    string
		expected int64
	}
	testCases := []testCase{
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "invalid",
			input:    "asfkhayakuwmrawgd214512",
			expected: 0,
		},
		{
			name:     "first",
			input:    "0-0",
			expected: 0,
		},
		{
			name:     "happy 1",
			input:    "1692632086370-0",
			expected: 1692632086370,
		},
		{
			name:     "happy",
			input:    "1692632086370-15",
			expected: 1692632086370,
		},
		{
			name:     "max",
			input:    fmt.Sprintf("%d-%d", math.MaxInt64, math.MaxInt64),
			expected: math.MaxInt64,
		},
	}
	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actual := parseRedisMessageID(tt.input)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
