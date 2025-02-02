package openapiapp

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/yanakipe/bot/internal/logger"
)

// An http.Handler that reroutes request with given prefix to a different handler, with a
// different prefix.
type RewriteRequestHandler struct {
	prefixFrom string
	prefixTo   string
	handlerTo  http.Handler
}

func (h RewriteRequestHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	after, found := strings.CutPrefix(request.URL.Path, h.prefixFrom)
	if !found {
		http.NotFound(writer, request)
		return
	}

	oldPath := request.URL.Path
	newPath := fmt.Sprintf("%s%s", h.prefixTo, after)
	logger.Debug(
		ctx,
		fmt.Sprintf("rewriting request %q to %q", oldPath, newPath),
	)
	request.URL.Path = newPath

	h.handlerTo.ServeHTTP(writer, request)
}

// Route requests starting with `prefixFrom` to `handlerTo`, rewriting the prefix to
// `prefixTo`. `prefixFrom` is passed as the pattern to ServerMux.Handle, so use a
// trailing slash to match a subtree. For example:
//
//	AddPrefixHandler(
//	  muxFrom, "/foo/",
//	  barHandler, "/bar/"
//	)
func AddPrefixHandler(
	muxFrom *http.ServeMux,
	prefixFrom string,
	handlerTo http.Handler,
	prefixTo string,
) {
	muxFrom.Handle(
		prefixFrom,
		RewriteRequestHandler{
			prefixFrom: prefixFrom,
			handlerTo:  handlerTo,
			prefixTo:   prefixTo,
		},
	)
}
