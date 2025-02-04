package resttooling

import (
	"fmt"
	"net/http"

	"github.com/yanakipre/bot/internal/logger"
)

type RouteNameFunc func(r *http.Request) UrlMethod

type UrlMethod struct {
	OperationID string
	Url         string
	Method      string
}

type routeT interface {
	Name() string
}

type serverT[R routeT] interface {
	FindRoute(string, string) (R, bool)
}

func UrlMethodGetter[R routeT](
	h serverT[R],
	unknownOperation string,
) func(*http.Request) UrlMethod {
	return func(r *http.Request) UrlMethod {
		if route, found := h.FindRoute(r.Method, r.URL.Path); !found {
			logger.Error(
				r.Context(),
				fmt.Sprintf("could not find route %s %s", r.Method, r.URL.Path),
			)
			return UrlMethod{Method: r.Method, OperationID: unknownOperation}
		} else {
			return UrlMethod{
				OperationID: route.Name(),
				Method:      r.Method,
			}
		}
	}
}

// Hardcoded as asap solution. Move to configs if needed
var UrlStopList = []string{
	"/cable",
	"/healthz",
	"/metrics",
}
