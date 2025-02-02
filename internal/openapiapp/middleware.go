package openapiapp

import "net/http"

type Middleware func(next http.Handler) http.Handler

// Middlewares is a sugar to skip zero middlewares
// AND have a full list of available middlewares in one place.
func Middlewares(mw ...Middleware) []Middleware {
	r := make([]Middleware, 0, len(mw))
	for i := range mw {
		if mw[i] == nil {
			continue
		}
		r = append(r, mw[i])
	}
	return r
}

// Wrap http handler with middlewares.
func Wrap(handler http.Handler, mw ...Middleware) http.Handler {
	resultHandler := handler
	for _, middleware := range mw {
		resultHandler = middleware(resultHandler)
	}
	return resultHandler
}
