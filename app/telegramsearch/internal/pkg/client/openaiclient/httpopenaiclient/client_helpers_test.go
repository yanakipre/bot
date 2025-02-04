package httpopenaiclient

import (
	"net/http"
	"strings"
	"testing"

	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"

	"github.com/yanakipre/bot/internal/secret"
)

var excludeReqHeaders = []string{
	"X-Request-Id",
	"Authorization",
}

var excludeRespHeaders = []string{
	"Server",
	"Report-To",
	"Retry-After",
	"Cf-Ray",
	"Cf-Cache-Status",
	"Openai-Organization",
	"Set-Cookie",
	"X-Request-Id",
	"Openai-Processing-Ms",
}

func fixturePath(name string) string {
	return strings.Replace(name, " ", "_", -1)
}

func clientWithTransport(
	t *testing.T,
	ttName string,
	recMode recorder.Mode,
	realTransport http.RoundTripper,
) (cancel func(), client *http.Client) {
	r, err := recorder.NewWithOptions(&recorder.Options{
		CassetteName:       fixturePath("internal/fixtures/" + t.Name() + "/" + ttName),
		RealTransport:      realTransport,
		Mode:               recMode,
		SkipRequestLatency: true,
	})
	if err != nil {
		panic(err)
	}
	cancel = func() { _ = r.Stop() } // Make sure recorder is stopped once done with it

	client = r.GetDefaultClient()

	// Add a hook which removes Authorization headers from all requests
	hook := func(i *cassette.Interaction) error {
		for _, header := range excludeReqHeaders {
			delete(i.Request.Headers, header)
		}
		for _, header := range excludeRespHeaders {
			delete(i.Response.Headers, header)
		}
		return nil
	}
	r.AddHook(hook, recorder.AfterCaptureHook)
	return cancel, client
}

func ClientWithConfigAndRecorder(
	t *testing.T,
	testName string,
	cfg Config,
) (cancel func(), client *Client) {
	recorderMode := recorder.ModeReplayOnly
	if cfg.ApiKey.Unmask() != "" {
		recorderMode = recorder.ModeRecordOnly
	}

	cancel, httpClient := clientWithTransport(
		t,
		testName,
		recorderMode,
		defaultHTTPClient(cfg).Transport,
	)
	cfg.httpClient = httpClient
	return cancel, NewClient(cfg)
}

func ClientWithRecorder(
	t *testing.T,
	testName string,
	// TODO: maybe it's better to use pointer to accessKey and use if not nil
	cfg Config,
	accessKey secret.String,
) (cancel func(), client *Client) {
	cfg.ApiKey = accessKey
	return ClientWithConfigAndRecorder(t, testName, cfg)
}
