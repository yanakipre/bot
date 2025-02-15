// Package networkerrs provides tooling to classify errors as temporary.
// It is used by the "presenters" layer to determine if a network error is temporary and is OK to retry.
// The reason it's exposed for imports is because we have a number of APIs
// that don't rely on the /internal/status package yet. And they must present errors in the same way.
package networkerrs

import (
	"errors"
	"syscall"
)

type (
	timeoutInterface interface{ Timeout() bool }
)

// isConnectionResetByPeer signals of "read: connection reset by peer" error.
// In other words, client received a RST packet from the server.
//
// That might mean many things, some of which are:
//  1. intermediate proxy broke the connection in flight.
//  2. network errors on switches, etc.
//  3. application acted in a wrong way, completely committing the successful request to it's DB,
//     but could not respond correctly, abruptly breaking TCP connection.
func isConnectionResetByPeer(err error) bool {
	return errors.Is(err, syscall.ECONNRESET)
}

// isConnectionRefused checks if the given error is a "connection refused" error.
//
// This function is useful for determining if a network error is due to a connection
// being refused, which can occur if the server is not accepting connections or if
// there is a network issue preventing the connection from being established.
// We believe that all the servers we're trying to connect to will eventually become online,
// so we consider "connection refused" temporary and safe to retry without thinking of HTTP request semantics.
// Because the request has never reached the server yet.
func isConnectionRefused(err error) bool {
	return errors.Is(err, syscall.ECONNREFUSED)
}

func IsNetworkErrTranslatesToUnavailable(err error) bool {
	if isConnectionResetByPeer(err) {
		return true
	}
	if isConnectionRefused(err) {
		return true
	}
	var timeoutErr timeoutInterface
	if errors.As(err, &timeoutErr) {
		return timeoutErr.Timeout()
	}
	return false
}
