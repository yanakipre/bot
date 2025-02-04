package worker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetWorkerIndex(t *testing.T) {
	ctx := context.Background()
	type args struct {
		hostname string
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		{
			name: "in k8s",
			args: args{hostname: "yanakipre-console-api-1"},
			want: 1,
		},
		{
			name: "locally",
			args: args{hostname: "console.local"},
			want: 0,
		},
		{
			name: "locally, that looks like the k8s format",
			args: args{hostname: "foo-bar"},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetWorkerIndex(ctx, tt.args.hostname)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
