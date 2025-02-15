package networkerrs

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_isConnectionResetByPeer(t *testing.T) {
	// in ipv4 there is no official non-routable address. So we try our best.
	// and there might be no support for ipv6, so we can't use it first.
	possibleNonRoutableAddresses := []string{
		"0.0.0.0:65535",
		"127.0.0.1:65535",   // localhost, not reachable
		"192.168.0.1:65535", // private IP address, not reachable
		"::1:65535",         // IPv6 localhost, not reachable
		"2001:0db8:85a3:0000:0000:8a2e:0370:7334:65535", // IPv6 loopback, not reachable
	}
	for _, guess := range possibleNonRoutableAddresses {
		_, err := net.Dial("tcp", guess)
		if err != nil {
			require.True(t, isConnectionRefused(err))
			return
		}
	}
	t.Fatal("managed to connect to all the test servers, in theory possible, but in practice this should not happen")
}
