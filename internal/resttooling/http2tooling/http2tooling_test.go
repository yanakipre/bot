package http2tooling

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnsureGetBodyMethod(t *testing.T) {
	const expectedBody = "test request body"

	req, err := http.NewRequest(http.MethodPost, "http://example.com", strings.NewReader(expectedBody))
	require.NoError(t, err, "Failed to create request")

	req.GetBody = nil

	// Ensure GetBody is set
	require.NoError(t, EnsureGetBodyMethod(req), "EnsureGetBodyMethod failed")

	newBody, err := req.GetBody()
	require.NoError(t, err, "Failed to get new body")
	gotBody, err := io.ReadAll(newBody)
	require.NoError(t, err, "Failed to read from GetBody")
	require.Equal(t, expectedBody, string(gotBody), "GetBody content mismatch")

	// Test that original Body can still be read
	originalBody, err := io.ReadAll(req.Body)
	require.NoError(t, err, "Failed to read from Body")
	require.Equal(t, expectedBody, string(originalBody), "Body content mismatch")
}
