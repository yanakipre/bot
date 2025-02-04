package openapiapp_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/openapiapp"
	"github.com/yanakipre/bot/internal/testtooling"
	"github.com/yanakipre/bot/internal/timer"
)

func TestGracefulShutdown(t *testing.T) {
	testtooling.SkipShort(t)
	// this test checks graceful shutdown
	// it is needed to ensure the contexts are used correctly because we messs with server.BaseContext
	// it creates a fake openapi.App with a fake healthcheck route with sleep
	// then it sends a request and cancels the context and initates the server shtudown while the server is sleeping before response
	// the expected behaviour is that we should still get the expected response after sleeping and the handler context is not cancelled
	logger.SetNewGlobalLoggerQuietly(logger.DefaultConfig())
	appCtx, cancel := context.WithCancel(context.Background())
	requestGot := make(chan bool)
	app := openapiapp.New(
		openapiapp.Config{Addr: "0.0.0.0:1111", ReadHeaderTimeout: time.Second},
		http.DefaultServeMux,
		[]openapiapp.Middleware{},
		func(ctx context.Context) (any, error) {
			requestGot <- true
			timeout := time.NewTimer(3 * time.Second)
			defer timer.StopTimer(timeout)
			select {
			case <-ctx.Done(): // this is important, check that no one canceled the context while we sleep
				return map[string]string{}, fmt.Errorf("cancelled")
			case <-timeout.C:
				return map[string]string{"response": "after sleep"}, nil
			}
		},
	)
	go app.StartServer(appCtx)

	// setup client
	type response struct {
		r   *http.Response
		err error
	}
	responseChan := make(chan response)
	go func() {
		time.Sleep(500 * time.Millisecond) // give server time to start serving
		requestCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		request, err := http.NewRequestWithContext(requestCtx, "GET", "http://127.0.0.1:1111/healthz", nil)
		if err != nil {
			panic("failed to build request")
		}
		r, err := http.DefaultClient.Do(request)
		responseChan <- response{r, err}
	}()
	timeout := time.NewTimer(1 * time.Second)
	defer timer.StopTimer(timeout)
	select {
	case <-requestGot:
	case <-timeout.C:
		panic("server has not got the request")
	}
	// server started processing the request, now cancel and shutdown
	cancel()
	go app.ShutdownServer(context.Background())

	// wait for reply
	timeout.Reset(5 * time.Second)
	var r response
	select {
	case r = <-responseChan:
	case <-timeout.C:
		panic("no reply")
	}

	// check response
	assert.NotNil(t, r.r)
	assert.Nil(t, r.err)
	assert.Equal(t, r.r.StatusCode, 200)
	body, err := io.ReadAll(r.r.Body)
	assert.Nil(t, err)
	respMap := make(map[string]string)
	err = json.Unmarshal(body, &respMap)
	assert.Nil(t, err)
	assert.Equal(t, map[string]string{"response": "after sleep"}, respMap)
}
