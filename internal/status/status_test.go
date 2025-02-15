package status

import (
	"errors"
	"fmt"
	"testing"

	"github.com/yanakipre/bot/internal/semerr"
)

func TestStatus_Message(t *testing.T) {
	tests := map[string]struct {
		err  error
		want string
	}{
		"unknown error": {
			err:  errors.New("foo"),
			want: "unknown error",
		},
		"semantic error": {
			err:  semerr.NotFound("not found"),
			want: "not found",
		},
		"wrapped semantic error": {
			err:  fmt.Errorf("oh no: %w", semerr.ResourceLocked("foo")),
			want: "foo",
		},
		"using rawError": {
			err:  rawError{err: errors.New("foo")},
			want: "foo",
		},
	}
	for name, tt := range tests {
		t.Run(
			name, func(t *testing.T) {
				st := FromError(tt.err)
				if got := st.Message(); got != tt.want {
					t.Errorf("Message() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
