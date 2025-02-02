package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/yanakipe/bot/internal/encodingtooling"
)

func Test_genConfig(t *testing.T) {
	type args struct {
		cfg any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "duration",
			args: args{
				DurationTestSingleField{
					LoadDuration: encodingtooling.Duration{Duration: time.Minute * 3},
				},
			},
			want: "timeout: 3m0s\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := genConfig(tt.args.cfg)
			require.Equal(t, tt.want, string(got))
		})
	}
}
