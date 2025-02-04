package resttooling

import (
	"context"
	"errors"
	"net/http"
)

type AuthInCtx string

const authMethodContextKey AuthInCtx = "yanakipreauthmethod"

type AuthMethod string

const (
	AuthMethodUNKNOWN AuthMethod = "UNKNOWN"
	AuthMethodCookie  AuthMethod = "cookie"
	AuthMethodToken   AuthMethod = "token"
	AuthMethodOAuth   AuthMethod = "oauth"
	AuthMethodJWT     AuthMethod = "jwt"
)

type AuthenticationMethod interface {
	// Auth controls authentication process.
	// To signal
	// * of unhandled error, return non-nil err.
	// * of SUCCESSFUL authentication, return non-nil stop. This will stop authentication process.
	// * of authentication that was NOT successful, return (nil, nil). This will continue auth
	//   process and try other methods if specified.
	Auth(w http.ResponseWriter, r *http.Request) (stop *http.Request, err error)
	GetName() string
}

var ErrAuthMethodNotInContext = errors.New("authentication method not in context")

// AuthMethodToContext sets authentication method that was used during authentication.
func AuthMethodToContext(ctx context.Context, method AuthMethod) context.Context {
	return context.WithValue(ctx, authMethodContextKey, method)
}

func AuthMethodFromContext(ctx context.Context) (AuthMethod, error) {
	m, ok := ctx.Value(authMethodContextKey).(AuthMethod)
	if !ok {
		return AuthMethodUNKNOWN, ErrAuthMethodNotInContext
	}
	return m, nil
}
