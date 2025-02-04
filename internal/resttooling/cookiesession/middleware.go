package cookiesession

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"

	"github.com/yanakipre/bot/internal/secret"
)

const secondsPerYear = 60 /*s*/ * 60 /*m*/ * 24 /*h*/ * 365 /*d*/

func CookieStoreToContext(ctx context.Context, store *sessions.CookieStore) context.Context {
	return context.WithValue(ctx, cookieStoreContextKey, store)
}

func NewStore(secret string) *sessions.CookieStore {
	store := sessions.NewCookieStore([]byte(secret))

	store.MaxAge(secondsPerYear)

	store.Options.HttpOnly = true
	store.Options.Secure = true

	return store
}

// Middleware sets valid cookie storage into context.
func Middleware(secret secret.String) func(next http.Handler) http.Handler {
	cookieStore := NewStore(secret.Unmask())

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*r = *r.WithContext(CookieStoreToContext(r.Context(), cookieStore))
			next.ServeHTTP(w, r)
		})
	}
}

func FromContext(ctx context.Context, r *http.Request) (*sessions.Session, error) {
	cs, ok := ctx.Value(cookieStoreContextKey).(*sessions.CookieStore)
	if !ok {
		return nil, fmt.Errorf("cookie store cannot be received from context")
	}
	session, err := cs.Get(r, cookieName)
	if err != nil {
		// user could break the cookie, but we can create clean session
		if session != nil && session.IsNew {
			return session, nil
		}
		return nil, fmt.Errorf("could not get valid cookie '%s' from store: %w", cookieName, err)
	}
	return session, nil
}

func ClearCookie(w http.ResponseWriter, r *http.Request) error {
	session, err := FromContext(r.Context(), r)
	if err != nil {
		return fmt.Errorf("could not get session from context: %w", err)
	}
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		return fmt.Errorf("could not save session: %w", err)
	}

	return nil
}
