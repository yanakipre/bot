package restretries

import (
	"context"
	"io"
	"net/http"

	"github.com/kamilsk/retry/v5"

	"github.com/yanakipe/bot/internal/resttooling/http2tooling"
)

var _ http.RoundTripper = &retryableRoundTripper{}

func (rt *retryableRoundTripper) Unwrap() http.RoundTripper { return rt.rt }

// RoundTrip implements http.RoundTripper with retries
func (rt *retryableRoundTripper) RoundTrip(originalReq *http.Request) (*http.Response, error) {
	var resp *http.Response
	ctx := originalReq.Context()

	// we must keep the original body
	// and substitute it on every retry
	// otherwise it is read only on first attempt.
	if err := http2tooling.EnsureGetBodyMethod(originalReq); err != nil {
		return nil, NewPermanentError(err)
	}

	var getBodyErr error

	action := func(_ context.Context) (err error) {
		r := originalReq

		if originalReq.GetBody != nil {
			r.Body, getBodyErr = r.GetBody()
			if getBodyErr != nil {
				return NewPermanentError(getBodyErr)
			}
		}
		if resp != nil {
			// get rid of response body from previous attempt
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}

		resp, err = rt.rt.RoundTrip(r)

		// wrap the error, if needed, and then return this
		// error for all requests that did not complete
		// successfully. The retry.How mechanism will abort
		// for all terminal mechanisms.
		if err = rt.checkNetworkError(err, r); err != nil {
			return err
		}

		// if the resquest was successful (from a network
		// perspective,) but returned an error code (e.g. >
		// 299), then maybe we should retry: this will create
		// a new error if needed.
		return rt.checkResponse(ctx, resp)
	}

	if err := retry.Do(ctx, action, rt.how...); err != nil {
		return resp, err
	}

	return resp, nil
}

// RetryByResponse manages retries logic based on valid http response.
// return nil to exit early.
// return any error to signal that retry would take place.
type RetryByResponse func(ctx context.Context, response *http.Response) error

// RetryByNetwork manages retries logic based on network error prior to valid response accepted by
// client.
// return nil to exit early.
// return any error to signal that retry would take place.
type RetryByNetwork func(err error, req *http.Request) error

// retryableRoundTripper implements retry logic
type retryableRoundTripper struct {
	how               retry.How
	checkResponse     RetryByResponse
	checkNetworkError RetryByNetwork
	rt                http.RoundTripper
}

func isRetryableRoundTripper(rt http.RoundTripper) bool {
	for {
		if _, ok := rt.(*retryableRoundTripper); ok {
			return true
		}

		if rt = unwrap(rt); rt == nil {
			return false
		}
	}
}

func unwrap(rt http.RoundTripper) http.RoundTripper {
	u, ok := rt.(interface{ Unwrap() http.RoundTripper })
	if !ok {
		return nil
	}
	return u.Unwrap()
}

// NewRetryableRoundTripper returns retryableRoundTripper with artificial cancellationContext.
// We do not have circuit breaker of any kind just now, so we just don't control this behavior in
// any way. If we want to introduce kamilsk/breaker, this cancellationContext should be changed to
// some circuit breaker.
func NewRetryableRoundTripper(
	rt http.RoundTripper,
	networkRetry RetryByNetwork,
	how retry.How,
	responseShouldBeRetried RetryByResponse,
) http.RoundTripper {
	if isRetryableRoundTripper(rt) {
		panic("cannot nest multiple retryable roundtrippers")
	}

	if networkRetry == nil {
		panic("must specify network retry")
	}

	if responseShouldBeRetried == nil {
		panic("must specify a response retry")
	}

	if len(how) == 0 {
		panic("must specify at least one retry strategy")
	}

	// make sure that the first strategy for retrying an error is
	// "do not retry terminal errors".
	how = append(retry.How{mustNotRetryTerminalErrorsStrategy}, how...)

	return &retryableRoundTripper{
		checkNetworkError: networkRetry,
		how:               how,
		checkResponse:     responseShouldBeRetried,
		rt:                rt,
	}
}
