package resttooling

import (
	"errors"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/semerr"
)

type ErrorHandler func(w http.ResponseWriter, r *http.Request, appErr error)

// RecoveryMiddleware recovers application from panics.
func RecoveryMiddleware(errorHandler ErrorHandler) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					logger.Panic(r.Context(), rvr)

					err := semerr.UnwrapPanic(rvr)

					// Check for a broken connection, as it is not really a
					// condition that warrants a panic stack trace.
					var ne *net.OpError
					if errors.As(err, &ne) {
						var se *os.SyscallError
						if errors.As(ne.Err, &se) {
							if strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
								strings.Contains(
									strings.ToLower(se.Error()),
									"connection reset by peer",
								) {

								// no need to response on broken pipe.
								return
							}
						}
					}

					if errors.Is(err, http.ErrAbortHandler) {
						// we don't recover http.ErrAbortHandler.
						// It means the client's request was aborted.
						return
					}

					errorHandler(w, r, semerr.UnwrapPanic(rvr))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
