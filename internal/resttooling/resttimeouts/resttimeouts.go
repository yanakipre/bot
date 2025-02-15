package resttimeouts

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/yanakipre/bot/internal/logger"
)

type Cfg struct {
	Timeout time.Duration
}

func New(rt http.RoundTripper, cfg Cfg) http.RoundTripper {
	return timeoutsRoundTripper{
		cfg: cfg,
		rt:  rt,
	}
}

// timeoutsRoundTripper implements a timeout for a single call of RoundTrip.
type timeoutsRoundTripper struct {
	cfg Cfg
	rt  http.RoundTripper
}

var errTemporary = errors.New("retryable rest timeout error")

func IsTimeout(err error) bool {
	return errors.Is(err, errTemporary)
}

func (t timeoutsRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	originalCtx := r.Context()
	ctx, cancel := context.WithTimeout(originalCtx, t.cfg.Timeout)
	defer cancel()

	sig := make(chan struct{})
	var retErr error
	var retResp *http.Response
	go func() {
		retResp, retErr = t.rt.RoundTrip(r)
		close(sig)
	}()

	select {
	case <-ctx.Done():
		logger.Info(ctx, "timeout for a single request")
		return nil, errors.Join(errTemporary, ctx.Err()) // suppress the error, because it will be retried
	case <-sig:
		return retResp, retErr
	}
}

var _ http.RoundTripper = (*timeoutsRoundTripper)(nil)
