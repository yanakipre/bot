package encodingtooling

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_camelToSnake(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"CreatedAt", "created_at"},
		{"", ""},
		{"already_snake", "already_snake"},
		{"A", "a"},
		{"AA", "aa"},
		{"AaAa", "aa_aa"},
		{"HTTPRequest", "http_request"},
		{"BatteryLifeValue", "battery_life_value"},
		{"Id0Value", "id0_value"},
		{"ID0Value", "id0_value"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			require.Equal(t, tt.want, CamelToSnake(tt.input))
		})
	}
}
