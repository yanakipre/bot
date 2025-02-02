package resttooling

import (
	"errors"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/yanakipe/bot/internal/semerr"
)

type ErrorHandler func(w http.ResponseWriter, r *http.Request, appErr error)

// RecoveryMiddleware recovers application from panics.
func RecoveryMiddleware(errorHandler ErrorHandler) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					// Check for a broken connection, as it is not really a
					// condition that warrants a panic stack trace.
					var brokenPipe bool
					var ne *net.OpError
					asError, ok := rvr.(error)
					if ok && errors.As(asError, &ne) {
						if se, ok := ne.Err.(*os.SyscallError); ok {
							if strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
								strings.Contains(
									strings.ToLower(se.Error()),
									"connection reset by peer",
								) {
								brokenPipe = true
							}
						}
					}

					if brokenPipe {
						// no need to response on broken pipe.
						return
					}

					if rvr == http.ErrAbortHandler {
						// we don't recover http.ErrAbortHandler.
						// It means the client's request was aborted.
						return
					}

					err, ok := rvr.(error)
					if !ok {
						errorHandler(w, r, semerr.Internal("internal error happened"))
						return
					}
					errorHandler(w, r, err)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
