package restretries

import (
	"context"
	"errors"
	"fmt"
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
	"github.com/yanakipre/bot/internal/testtooling"
)

func Test_retryableRoundTripper_RoundTrip(t *testing.T) {
	type fields struct {
		how                         retry.How
		responseShouldBeRetried     func() (r RetryByResponse, cnt *int)
		networkErrorShouldBeRetried func() (r RetryByNetwork, cnt *int)
		rt                          http.RoundTripper
	}
	testtooling.SetNewGlobalLoggerQuietly()
	type args struct {
		originalReq func() *http.Request
	}
	tests := []struct {
		name                 string
		fields               fields
		countNetworkRetries  int
		countResponseRetries int
		args                 args
		want                 *http.Response
		wantErr              bool
	}{
		{
			name: "response error with body and 1 retry attempt",
			fields: fields{
				how: retry.How{
					mustNotRetryTerminalErrorsStrategy,
					strategy.Limit(3),
				},
				responseShouldBeRetried: func() (r RetryByResponse, cnt *int) {
					counter := 0
					cnt = &counter
					return func(ctx context.Context, response *http.Response) error {
						counter += 1
						if counter == 2 {
							logger.Warn(ctx, "permanent", zap.Int("count", counter))
							return NewPermanentError(fmt.Errorf("test-response-error"))
						}
						logger.Warn(ctx, "retry temporary", zap.Int("count", counter))
						return fmt.Errorf("should be retried")
					}, cnt
				},
				networkErrorShouldBeRetried: func() (r RetryByNetwork, cnt *int) {
					counter := 0
					cnt = &counter
					return func(err error, _ *http.Request) error {
						counter++
						if err != nil {
							t.Errorf("error should be nil: %v", err)
						}
						return nil
					}, cnt
				},
				rt: http.DefaultTransport,
			},
			countNetworkRetries:  2, // should call the retry function every time
			countResponseRetries: 2,
			args: args{
				originalReq: func() *http.Request {
					r, err := http.NewRequest(
						"GET",
						"http://127.0.0.1:8090/",
						strings.NewReader("body"),
					)
					if err != nil {
						panic(err)
					}
					return r
				},
			},
			wantErr: true,
		},
		{
			name: "network error when no body and no retries",
			fields: fields{
				how: retry.How{
					mustNotRetryTerminalErrorsStrategy,
					strategy.Limit(3),
				},
				networkErrorShouldBeRetried: func() (r RetryByNetwork, cnt *int) {
					counter := 0
					cnt = &counter
					return func(err error, req *http.Request) error {
						counter += 1
						logger.Warn(context.Background(), "permanent", zap.Int("count", counter))
						return NewPermanentError(fmt.Errorf("test-error"))
					}, cnt
				},
				rt: http.DefaultTransport,
			},
			countNetworkRetries: 1,
			args: args{
				originalReq: func() *http.Request {
					r, err := http.NewRequest("GET", "https://non-existing-test/", nil)
					if err != nil {
						panic(err)
					}
					return r
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//
			// start http server on 8090 to connect to
			// it just accepts http sessions, that's it. No HTTP handlers.
			//
			ctx := context.Background()
			server := &http.Server{Addr: ":8090", Handler: nil}
			go func() {
				if err := server.ListenAndServe(); err != http.ErrServerClosed {
					panic(err)
				}
			}()
			// wait till ready
			ok := false
			for i := 0; i < 5; i++ {
				_, err := net.Dial("tcp", "127.0.0.1:8090")
				if err != nil {
					time.Sleep(time.Second)
					continue
				}
				ok = true
				break
			}
			require.True(t, ok, "server unavailable")
			defer func() { _ = server.Shutdown(ctx) }()

			//
			// Create roundtripper to test
			//
			var networkErrorShouldBeRetried RetryByNetwork
			var countedNetworkRetries *int
			if tt.fields.networkErrorShouldBeRetried != nil {
				networkErrorShouldBeRetried, countedNetworkRetries = tt.fields.networkErrorShouldBeRetried()
			}

			var responseShouldBeRetried RetryByResponse
			var countedResponseRetries *int
			if tt.fields.responseShouldBeRetried != nil {
				responseShouldBeRetried, countedResponseRetries = tt.fields.responseShouldBeRetried()
			}

			rt := &retryableRoundTripper{
				how:               tt.fields.how,
				checkResponse:     responseShouldBeRetried,
				checkNetworkError: networkErrorShouldBeRetried,
				rt:                tt.fields.rt,
			}

			//
			// Run test
			//
			got, err := rt.RoundTrip(tt.args.originalReq())

			if tt.wantErr {
				require.Error(t, err)
				require.True(t, IsPermanentError(err), "not permanent error, got '%s'", err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}

			if countedNetworkRetries != nil {
				require.Equal(t, tt.countNetworkRetries, *countedNetworkRetries)
			}
			if countedResponseRetries != nil {
				require.Equal(t, tt.countResponseRetries, *countedResponseRetries)
			}
		})
	}
}

func TestMustNotRetryTerminalError(t *testing.T) {
	ctx := context.Background()
	t.Run("LimitRespected", func(t *testing.T) {
		var callCount uint
		err := retry.Do(ctx,
			func(_ctx context.Context) error {
				callCount++
				return nil
			},
			mustNotRetryTerminalErrorsStrategy,
			strategy.Limit(10),
		)
		if err != nil {
			t.Fatal(err)
		}
		if callCount != 1 {
			t.Fatal("did not call function:", callCount)
		}
	})
	t.Run("RetriesError", func(t *testing.T) {
		const errAfter = 2

		var callCount uint
		err := retry.Do(ctx,
			func(_ctx context.Context) error {
				callCount++
				if callCount < errAfter {
					return errors.New("foo")
				}
				return nil
			},
			mustNotRetryTerminalErrorsStrategy,
			strategy.Limit(10),
		)
		if callCount != errAfter {
			t.Fatalf("called function %d, ", callCount)
		}

		if err != nil {
			t.Fatal("unexpected error", err)
		}
	})
	t.Run("EventualTerminalError", func(t *testing.T) {
		const errAfter = 5

		var callCount uint
		err := retry.Do(ctx,
			func(_ctx context.Context) error {
				callCount++
				if callCount < errAfter {
					return errors.New("foo")
				}
				return NewPermanentError(errors.New("bar"))
			},
			mustNotRetryTerminalErrorsStrategy,
			strategy.Limit(10),
		)
		if callCount != errAfter {
			t.Fatalf("called function %d times, not %d", callCount, errAfter)
		}

		if !IsPermanentError(err) {
			t.Fatal("should have been permanent error", err)
		}
	})
	t.Run("EventualSucceeds", func(t *testing.T) {
		const errAfter = 100

		var callCount uint
		err := retry.Do(ctx,
			func(_ctx context.Context) error {
				callCount++
				if callCount < errAfter {
					return errors.New("foo")
				}
				return NewPermanentError(errors.New("bar"))
			},
			mustNotRetryTerminalErrorsStrategy,
			strategy.Limit(10),
		)
		if callCount != 10 {
			t.Fatalf("should have retried until limit expired, not %d", callCount)
		}

		if IsPermanentError(err) {
			t.Fatal("should NOT have been permanent error", err)
		}
	})
}
