package resttooling

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/yanakipre/bot/internal/secret"
	"github.com/yanakipre/bot/internal/semerr"
)

const (
	AuthorizationHeader = "Authorization"
	AuthPrefix          = "Bearer "
	UnknownAddr         = "UNKNOWN"
	invalidAuthMsg      = "not authenticated"
)

// SetBearerToken sets a bearer token into outgoing request.
func SetBearerToken(token string, req *http.Request) {
	req.Header.Set(AuthorizationHeader, BearerPayload(token))
}

func BearerPayload(token string) string {
	return fmt.Sprintf("%s%s", AuthPrefix, token)
}

var ErrNoAuthHeader = errors.New("no auth header supplied")

var loweredAuthPrefix = strings.ToLower(AuthPrefix)

// getBearerTokenFromReq returns token value from http request
func getBearerTokenFromReq(req *http.Request) (secret.String, error) {
	authHdr := req.Header.Get(AuthorizationHeader)
	// Check for the Authorization header.
	if authHdr == "" {
		return secret.String{}, ErrNoAuthHeader
	}
	// We expect a header value of the form "Bearer <token>", with 1 space after
	// Bearer, per spec.
	if len(authHdr) <= len(loweredAuthPrefix) {
		return secret.String{}, errors.New("header is too short")
	}
	if strings.ToLower(authHdr[:len(loweredAuthPrefix)]) != loweredAuthPrefix {
		return secret.String{}, errors.New("header is malformed")
	}
	return secret.NewString(authHdr[len(loweredAuthPrefix):]), nil
}

func getSourceAddr(req *http.Request) string {
	addr := req.Header.Get("X-Forwarded-For")
	if addr == "" {
		addr = UnknownAddr
	}
	return addr
}

type BearerTokenReq struct {
	TokenFromReq secret.String
	SourceAddr   string
}

// BearerTokenReqFromHTTP returns InvalidInput semerr if token missing or invalid
func BearerTokenReqFromHTTP(r *http.Request) (BearerTokenReq, error) {
	tokenFromReq, err := getBearerTokenFromReq(r)
	if err != nil {
		return BearerTokenReq{}, semerr.WrapWithInvalidInput(err, invalidAuthMsg)
	}
	return BearerTokenReq{
		TokenFromReq: tokenFromReq,
		SourceAddr:   getSourceAddr(r),
	}, nil
}
