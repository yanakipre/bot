package requestid

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFromEmptyContext(t *testing.T) {
	ctx := context.Background()
	rid, ok := FromContext(ctx)
	require.False(t, ok)
	require.Empty(t, rid)

	rid = FromContextOrNew(ctx)
	require.NotEmpty(t, rid)
}

func TestToContext(t *testing.T) {
	ctx := context.Background()
	rid := FromContextOrNew(ctx)
	ctxWithRID := WithRequestID(ctx, rid)
	require.NotNil(t, ctxWithRID)
	require.NotEqual(t, ctx, ctxWithRID)

	ridFromCtx, ok := FromContext(ctxWithRID)
	require.True(t, ok)
	require.Equal(t, rid, ridFromCtx)

	ridFromCtx = FromContextOrNew(ctxWithRID)
	require.Equal(t, rid, ridFromCtx)
}

func TestRequestIDFromContextOrNew(t *testing.T) {
	type args struct {
		r func() *http.Request
	}
	tests := []struct {
		name          string
		args          args
		wantReqHeader string
		wantId        string
	}{
		{
			name: "already in context",
			args: args{
				r: func() *http.Request {
					r, err := http.NewRequest("", "", nil)
					if err != nil {
						panic(err)
					}
					r = r.WithContext(WithRequestID(context.Background(), "test-r-id"))
					return r
				},
			},
			wantReqHeader: "",
			wantId:        "test-r-id",
		},
		{
			name: "sent in header",
			args: args{
				r: func() *http.Request {
					r, err := http.NewRequest("", "", nil)
					if err != nil {
						panic(err)
					}
					r.Header.Add("x-request-id", "test-r-id")
					return r
				},
			},
			wantReqHeader: "test-r-id",
			wantId:        "test-r-id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.args.r()
			rID := RequestIDFromRequestOrNew(req)
			require.Equal(t, tt.wantId, rID)
			require.Equal(t, tt.wantReqHeader, req.Header.Get("x-request-id"))
		})
	}

	// idempotency check
	r, err := http.NewRequest("", "", nil)
	if err != nil {
		panic(err)
	}
	rID := RequestIDFromRequestOrNew(r)
	require.NotEqual(t, rID, "")
	rID2 := RequestIDFromRequestOrNew(r)
	require.Equal(t, rID, rID2)
}
