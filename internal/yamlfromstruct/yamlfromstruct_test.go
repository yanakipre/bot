package yamlfromstruct

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yanakipre/bot/internal/testtooling"
)

func TestGenerate(t *testing.T) {
	t.Parallel()
	type CamelCaseTest struct {
		FieldName string
	}

	ctx := context.Background()
	type args struct {
		ctx context.Context
		s   any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test camelcase",
			args: args{
				ctx: ctx,
				s:   CamelCaseTest{},
			},
			want: `field_name: ""
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testtooling.SetNewGlobalLoggerQuietly()
			got := Generate(tt.args.ctx, tt.args.s)
			require.Equal(t, tt.want, got)
		})
	}
}
