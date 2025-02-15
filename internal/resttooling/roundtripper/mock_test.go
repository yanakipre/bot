package roundtripper

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yanakipre/bot/internal/logger"
)

func TestMockRoundTripper(t *testing.T) {
	logger.SetNewGlobalLoggerQuietly(logger.DefaultConfig())

	mux := http.NewServeMux()
	called := 0
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		called += 1
		w.WriteHeader(http.StatusNoContent)
	})

	uri, err := url.Parse("http://mock.localhost/")
	require.Nil(t, err)

	req := http.Request{
		Method: "GET",
		URL:    uri,
	}

	rt := MockRoundTripper(mux)
	res, err := rt.RoundTrip(&req)

	require.Nil(t, err)
	assert.Equal(t, http.StatusNoContent, res.StatusCode)
	assert.Equal(t, 1, called)
}
