package resttooling

import (
	"errors"
	"net/http"

	"github.com/yanakipre/bot/internal/semerr"
)

var _ AuthenticationMethod = ForbiddenRequestAuthentication{}

// ForbiddenRequestAuthentication responses with Authentication error
type ForbiddenRequestAuthentication struct{}

var errForbiddenRequestAuthentication = errors.New("ForbiddenRequestAuthentication")

func (ForbiddenRequestAuthentication) Auth(http.ResponseWriter, *http.Request) (*http.Request, error) {
	return nil, semerr.WrapWithAuthentication(
		errForbiddenRequestAuthentication,
		"supplied credentials do not pass authentication",
	)
}

func (ForbiddenRequestAuthentication) GetName() string {
	return "forbidden request authentication"
}
