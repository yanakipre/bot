package semerr

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var stack1 = callMeToo(
	callMeFirst(),
)

func callMeFirst() *stack {
	return callers()
}

func callMeToo(stack *stack) *stack {
	return stack
}

func Test_stack_Format(t *testing.T) {
	type args struct {
		verb string
	}
	tests := []struct {
		name string
		s    *stack
		args args
		want string
	}{
		{
			name: "happy path",
			s:    stack1,
			args: args{
				verb: "%+v",
			},
			want: "bot/internal/semerr/stack_test.go:11",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Contains(t, fmt.Sprintf(tt.args.verb, tt.s.StackTrace()), tt.want)
		})
	}
}
