// Package cookiesession
// Provides routines to manage user session stored in cookie.
package cookiesession

import (
	"github.com/gorilla/sessions"
)

type cookieStoreContext string

const (
	cookieStoreContextKey = cookieStoreContext("cookieStore")
	cookieName            = "zenith"
)

// UserIDContextKey represents id of authenticated user.
const UserIDContextKey = "user_id"

type CookieSession sessions.Session
