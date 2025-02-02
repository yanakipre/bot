package semerr

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
)

type ExpectFormat struct {
	ExpectString string
}

func TestFormatting(t *testing.T) {
	t.Parallel()
	type TestCase struct {
		name   string
		input  func(t *testing.T) error
		expect ExpectFormat
	}
	testCases := []TestCase{
		{
			name: "xerrors is compatible with semerr",
			input: func(t *testing.T) error {
				err := NotFound("not found")
				return xerrors.Errorf("text: %w", err)
			},
			expect: ExpectFormat{
				ExpectString: "text: not found",
			},
		},
		{
			name: "multiple wrap",
			input: func(t *testing.T) error {
				err := io.EOF
				err = WrapWithUnavailable(err, "unavailable")
				err = WrapWithNotImplemented(err, "not implemented")
				return WrapWithNotFound(err, "not found")
			},
			expect: ExpectFormat{
				ExpectString: "not found: not implemented: unavailable: EOF",
			},
		},
		{
			name: "single 'not found'",
			input: func(t *testing.T) error {
				return NotFound("not found")
			},
			expect: ExpectFormat{
				ExpectString: "not found",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.input(t)
			gotString := err.Error()
			require.Equal(t, tc.expect.ExpectString, gotString)
		})
	}
}
