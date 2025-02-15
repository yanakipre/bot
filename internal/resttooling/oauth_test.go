package resttooling

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_getOAuthTokenFromReq(t *testing.T) {
	t.Parallel()
	type args struct {
		req *http.Request
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
		want    string
	}{
		{
			name: "header with lowercase Bearer given",
			args: args{req: &http.Request{
				Header: map[string][]string{
					"Authorization": {"bearer token"},
				},
			}},
			want: "token",
		},
		{
			name: "header given",
			args: args{req: &http.Request{
				Header: map[string][]string{
					"Authorization": {"Bearer token"},
				},
			}},
			want: "token",
		},
		{
			name:    "no header given",
			args:    args{req: &http.Request{}},
			wantErr: ErrNoAuthHeader,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := getBearerTokenFromReq(tt.args.req)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, token.Unmask())
		})
	}
}
