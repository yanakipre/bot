package resttooling

import "net/http"

func SecurityHeadersHandler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			WithSecurityHeaders(w.Header().Add)
			next.ServeHTTP(w, r)
		})
	}
}

func WithSecurityHeaders(addHeader func(key, value string)) {
	addHeader("X-Frame-Options", "SAMEORIGIN")
	addHeader("X-Content-Type-Options", "nosniff")
	addHeader("Referrer-Policy", "strict-origin-when-cross-origin")
}
