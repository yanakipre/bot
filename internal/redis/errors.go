package redis

import "github.com/dranikpg/gtrs"

//
// redis-streams related errors below:
//

type (
	StreamParseError = gtrs.ParseError
	StreamAckError   = gtrs.AckError
	// StreamReadError signals either read or write operation failed.
	// YES, write operations too. See stream.go, "Add"ing to the stream might result in StreamReadError.
	StreamReadError = gtrs.ReadError
)
