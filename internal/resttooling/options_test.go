package resttooling

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/kamilsk/retry/v5"
	"github.com/kamilsk/retry/v5/strategy"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/resttooling/restretries"
	"github.com/yanakipre/bot/internal/testtooling"
)

type server struct {
	port     int
	addr     string
	shutdown func()
}

func neverRespondingHandlerFunc(w http.ResponseWriter, r *http.Request) {
	logger.Info(r.Context(), "accepted the connection, but never responding")
	time.Sleep(1000 * time.Second)
}

func respondingHandlerFunc(latency time.Duration, latCount int) http.HandlerFunc {
	respCount := -1
	return func(w http.ResponseWriter, r *http.Request) {
		respCount += 1
		if respCount < latCount {
			logger.Info(r.Context(), "accepted the connection, will respond later", zap.Duration("first_resp_latency", latency), zap.Int("resp_count", respCount))
			time.Sleep(latency)
		} else {
			logger.Info(r.Context(), "accepted the connection, will respond immediately", zap.Int("resp_count", respCount))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

func getServer(lg logger.Logger, h http.HandlerFunc) server {
	r := server{}
	listening := make(chan struct{})

	go func() {
		ln, err := net.Listen("tcp", "127.0.0.1:0") // 0 means "choose any free port"
		if err != nil {
			panic(fmt.Errorf("error setting up listener: %w", err))
		}
		defer ln.Close()
		lg.Info("listening")
		r.port = ln.Addr().(*net.TCPAddr).Port
		r.addr = "http://" + ln.Addr().String()
		s := http.Server{}
		r.shutdown = func() {
			err := s.Shutdown(context.Background())
			if err != nil {
				panic(err)
			}
		}

		listening <- struct{}{}
		_ = http.Serve(ln, h)
	}()
	// wait for the server to start listening
	<-listening
	return r
}

func TestWithTimeout(t *testing.T) {
	ctx := context.Background()
	testtooling.SetNewGlobalLoggerQuietly()
	t.Parallel()

	t.Run("timeout for POST, simple", func(t *testing.T) {
		lg := logger.FromContext(ctx)
		s := getServer(lg, neverRespondingHandlerFunc)
		defer s.shutdown()

		c := WrapHTTPClientWithOptions(&http.Client{},
			WithTransport(DefaultTransportConfig()),
			WithTimeout(time.Millisecond*63),
		)
		r, err := http.NewRequestWithContext(ctx, "POST", s.addr, strings.NewReader("test data"))
		require.NoError(t, err)
		_, err = c.Do(r)
		require.Error(t, err)
		require.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("no timeout for POST, without retries", func(t *testing.T) {
		lg := logger.FromContext(ctx)
		s := getServer(lg, respondingHandlerFunc(2*time.Millisecond, 1))
		defer s.shutdown()

		c := WrapHTTPClientWithOptions(&http.Client{},
			WithTransport(DefaultTransportConfig()),
			WithTimeout(time.Millisecond*63),
		)
		r, err := http.NewRequestWithContext(ctx, "POST", s.addr, strings.NewReader("test data"))
		require.NoError(t, err)
		resp, err := c.Do(r)
		require.NoError(t, err)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, "OK", string(body))
	})

	t.Run("server is healthy, small latency (2 msec), one attempt out of 3 max", func(t *testing.T) {
		lg := logger.FromContext(ctx).Named(t.Name())
		s := getServer(lg, respondingHandlerFunc(2*time.Millisecond, 1))
		t.Cleanup(s.shutdown)

		attemptsTaken := 0

		c := WrapHTTPClientWithOptions(&http.Client{},
			WithTransport(DefaultTransportConfig()),
			WithTimeout(time.Millisecond*63),
			WithRetries(
				restretries.StraightforwardNetworkRetry(),
				restretries.RepeatRetriableStatusCodes(),
				retry.How{
					strategy.Limit(3),
					func(breaker retry.Breaker, u uint, err error) bool {
						lg.Info("attempt taken", zap.Uint("count", u))
						attemptsTaken += 1
						return true
					},
				},
			),
		)
		r, err := http.NewRequestWithContext(ctx, "POST", s.addr, strings.NewReader("test data"))
		require.NoError(t, err)
		resp, err := c.Do(r)
		require.NoError(t, err)
		require.Equal(t, 1, attemptsTaken)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, "OK", string(body))
	})
	t.Run("server is healthy, huge latency (173 msec), 2 attempts out of 3 max", func(t *testing.T) {
		lg := logger.FromContext(ctx).Named(t.Name())
		s := getServer(lg, respondingHandlerFunc(173*time.Millisecond, 1))
		t.Cleanup(s.shutdown)

		attemptsTaken := 0

		c := WrapHTTPClientWithOptions(&http.Client{},
			WithTransport(DefaultTransportConfig()),
			WithTimeout(time.Millisecond*63),
			WithRetries(
				restretries.StraightforwardNetworkRetry(),
				restretries.RepeatRetriableStatusCodes(),
				retry.How{
					strategy.Limit(3),
					func(breaker retry.Breaker, u uint, err error) bool {
						lg.Info("attempt taken", zap.Uint("count", u))
						attemptsTaken += 1
						return true
					},
				},
			),
		)
		r, err := http.NewRequestWithContext(ctx, "POST", s.addr, strings.NewReader("test data"))
		require.NoError(t, err)
		resp, err := c.Do(r)
		require.NoError(t, err)
		require.Equal(t, 2, attemptsTaken)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, "OK", string(body))
	})
	t.Run("server is healthy, huge latency (173 msec) for all requests, 3 attempts out of 3 max", func(t *testing.T) {
		lg := logger.FromContext(ctx).Named(t.Name())
		s := getServer(lg, respondingHandlerFunc(173*time.Millisecond, 3))
		t.Cleanup(s.shutdown)

		attemptsTaken := 0

		c := WrapHTTPClientWithOptions(&http.Client{},
			WithTransport(DefaultTransportConfig()),
			WithTimeout(time.Millisecond*63),
			WithRetries(
				restretries.StraightforwardNetworkRetry(),
				restretries.RepeatRetriableStatusCodes(),
				retry.How{
					strategy.Limit(3),
					func(breaker retry.Breaker, u uint, err error) bool {
						lg.Info("attempt taken", zap.Uint("count", u))
						attemptsTaken += 1
						return true
					},
				},
			),
		)
		r, err := http.NewRequestWithContext(ctx, "POST", s.addr, strings.NewReader("test data"))
		require.NoError(t, err)
		_, err = c.Do(r)
		require.Error(t, err)
		require.Equal(t, 3, attemptsTaken)
		require.ErrorIs(t, err, context.DeadlineExceeded)
	})
}
