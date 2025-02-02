package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_isStringExactOrChild(t *testing.T) {
	type args struct {
		root      string
		candidate string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "child candidate",
			args: args{
				root:      "root",
				candidate: "root.child",
			},
			want: true,
		},
		{
			name: "same length",
			args: args{
				root:      "root",
				candidate: "toor",
			},
			want: false,
		},
		{
			name: "similar candidate",
			args: args{
				root:      "root",
				candidate: "rootsimilar",
			},
			want: false,
		},
		{
			name: "exact candidate",
			args: args{
				root:      "root",
				candidate: "root",
			},
			want: true,
		},
		{
			name: "short candidate",
			args: args{
				root:      "root",
				candidate: "c",
			},
			want: false,
		},
		{
			name: "long candidate",
			args: args{
				root:      "root",
				candidate: "i_am_really_long",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, isStringExactOrChild(tt.args.root, tt.args.candidate))
		})
	}
}

func Test_filterLoggerNameByExactSubNameStartswith(t *testing.T) {
	type args struct {
		filters    []ExactSubnameFilter
		loggerName string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "b.gormFOO.a stays",
			args: args{
				filters:    []ExactSubnameFilter{{"gorm"}},
				loggerName: "b.gormFOO.a",
			},
			want: true,
		},
		{
			name: "b.gorm.a out",
			args: args{
				filters:    []ExactSubnameFilter{{"gorm"}},
				loggerName: "b.gorm.a",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterLoggerNameByExactSubNameStartswith(tt.args.filters, tt.args.loggerName); got != tt.want {
				t.Errorf("filterLoggerNameByExactSubNameStartswith() = %v, want %v", got, tt.want)
			}
		})
	}
}
