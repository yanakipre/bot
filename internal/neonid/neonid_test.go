package neonid

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type generatorArgs struct {
	extraPrefix string
}

func TestGenerateProjectID(t *testing.T) {
	type testData struct {
		name  string
		args  generatorArgs
		check func(*testing.T, []string)
	}
	tests := []testData{
		{
			name: "no digit",
			args: generatorArgs{""},
			check: func(t *testing.T, parts []string) {
				assert.Len(t, parts, 3)
				assert.Len(t, parts[2], 8)
				for _, char := range parts[2] {
					assert.Contains(t, "0123456789", string(char))
				}
			},
		},
		{
			name: "prefix a1",
			args: generatorArgs{"a1"},
			check: func(t *testing.T, parts []string) {
				assert.Len(t, parts, 3)
				assert.Len(t, parts[2], 8)
				for _, char := range parts[2] {
					assert.Contains(t, "0123456789", string(char))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := NewGenerator(tt.args.extraPrefix)
			res := g.GenerateProjectID()
			tt.check(t, strings.Split(res, "-"))
		})
	}
}

func TestGenerateBranchID(t *testing.T) {
	type testData struct {
		name  string
		args  generatorArgs
		check func(*testing.T, []string)
	}
	tests := []testData{
		{
			name: "no digit",
			args: generatorArgs{""},
			check: func(t *testing.T, parts []string) {
				require.Len(t, parts, 4)
				assert.Equal(t, "br", parts[0])
				assert.Len(t, parts[3], 8)
				for _, char := range parts[3] {
					assert.Contains(t, "0123456789", string(char))
				}
			},
		},
		{
			name: "prefix a1",
			args: generatorArgs{"a1"},
			check: func(t *testing.T, parts []string) {
				require.Len(t, parts, 4)
				assert.Equal(t, "br", parts[0])
				assert.Len(t, parts[3], 8)
				assert.Equal(t, "a1", parts[3][0:2])
				for _, char := range parts[3] {
					assert.Contains(t, "0123456789abcdefghijklmnopqrstuvwxyz", string(char))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := NewGenerator(tt.args.extraPrefix)
			res := g.GenerateBranchID()
			tt.check(t, strings.Split(res, "-"))
		})
	}
}

func TestGenerateEndpointID(t *testing.T) {
	type testData struct {
		name  string
		args  generatorArgs
		check func(*testing.T, []string)
	}
	tests := []testData{
		{
			name: "no digit",
			args: generatorArgs{""},
			check: func(t *testing.T, parts []string) {
				require.Len(t, parts, 4)
				assert.Equal(t, "ep", parts[0])
				assert.Len(t, parts[3], 8)
				for _, char := range parts[3] {
					assert.Contains(t, "0123456789", string(char))
				}
			},
		},
		{
			name: "prefix a1",
			args: generatorArgs{"a1"},
			check: func(t *testing.T, parts []string) {
				require.Len(t, parts, 4)
				assert.Equal(t, "ep", parts[0])
				assert.Len(t, parts[3], 8)
				assert.Equal(t, "a1", parts[3][0:2])
				for _, char := range parts[3] {
					assert.Contains(t, "0123456789abcdefghijklmnopqrstuvwxyz", string(char))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := NewGenerator(tt.args.extraPrefix)
			res := g.GenerateEndpointID()
			tt.check(t, strings.Split(res, "-"))
		})
	}
}
