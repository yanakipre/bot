package jsontooling

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStrictUnmarshall(t *testing.T) {
	type args struct {
		data []byte
		v    any
	}
	type s struct {
		A string `json:"a"`
	}
	tests := []struct {
		name    string
		args    args
		wantErr require.ErrorAssertionFunc
	}{
		{
			name: "happy path",
			args: args{
				data: []byte("{\n  \"a\": \"hello\",\n  \"b\": \"unknown field\"\n}"),
				v:    &s{},
			},
			wantErr: func(t require.TestingT, err error, i ...any) {
				require.Error(t, err)
				require.EqualErrorf(t, err, "json: unknown field \"b\"", "unequal")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := StrictUnmarshal(tt.args.data, tt.args.v)
			tt.wantErr(t, err)
		})
	}
}
