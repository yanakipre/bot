package codeerr

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yanakipre/bot/internal/semerr"
)

func TestAsCodeErr(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name   string
		args   args
		expect ErrorCode
	}{
		{
			name: "translates codeerr to Code attribute",
			args: args{
				err: Wrap(ProjectsLimitExceeded, semerr.NotFound("test not found")),
			},
			expect: ProjectsLimitExceeded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AsCodeErr(tt.args.err)
			require.Equal(t, tt.expect, got.code)
		})
	}
}
