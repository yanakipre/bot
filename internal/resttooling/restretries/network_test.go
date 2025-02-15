package restretries

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/testtooling"
)

func TestStraightforwardNetworkRetry(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testtooling.SetNewGlobalLoggerQuietly()
	lg := logger.FromContext(ctx)
	// This function creates a network error that is "connection reset by peer" and returns it.
	connResetByPeerError := func(ctx context.Context) error {
		listening := make(chan struct{})
		var port int

		go func() {
			ln, err := net.Listen("tcp", "127.0.0.1:0") // 0 means "choose any free port"
			if err != nil {
				panic(fmt.Errorf("error setting up listener: %w", err))
			}
			defer ln.Close()
			lg.Info("listening")
			port = ln.Addr().(*net.TCPAddr).Port
			listening <- struct{}{}

			conn, err := ln.Accept()
			if err != nil {
				panic(fmt.Errorf("error accepting connection: %w", err))
			}
			lg.Info("connection accepted, giving time to write")
			time.Sleep(40 * time.Millisecond)
			lg.Info("closing connection")
			// Simulate the "reset by peer" by closing the connection immediately
			conn.Close()
		}()
		<-listening

		req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://127.0.0.1:%d/", port), nil)
		if err != nil {
			return err
		}
		_, err = http.DefaultClient.Do(req)
		if err != nil {
			if !strings.Contains(err.Error(), "connection reset by peer") {
				panic(fmt.Errorf("unexpected error: %w", err))
			}
			return err
		}
		panic("no error")
	}

	// client received a RST packet from the server.
	//
	// That might mean many things, some of the,:
	//  1. intermediate proxy broke the connection in flight.
	//  2. network errors on switches, etc.
	//  3. application acted in a wrong way, completely committing the successful request to it's DB,
	//     but could not respond correctly, abruptly breaking TCP connection.
	for _, method := range []string{"GET", "HEAD", "OPTIONS"} {
		t.Run(fmt.Sprintf("%s retries connection reset by peer", method), func(t *testing.T) {
			// it's safe to retry idempotent requests, RFC 7230.
			// we retry this request
			connRstErr := connResetByPeerError(ctx)
			r, err := http.NewRequestWithContext(ctx, method, "http://yanakipre-fakeurl", nil)
			require.NoError(t, err)

			policy := StraightforwardNetworkRetry()
			policyDesicion := policy(connRstErr, r)
			// error means "retry this request"
			require.Error(t, policyDesicion)
			require.False(t, IsPermanentError(policyDesicion)) // we retry this request
		})
	}
	for _, method := range []string{"POST", "DELETE", "PUT"} {
		t.Run(fmt.Sprintf("%s not retries connection reset by peer", method), func(t *testing.T) {
			// it's NOT safe to retry non-idempotent requests, RFC 7230.
			// we do not retry this request
			connRstErr := connResetByPeerError(ctx)
			r, err := http.NewRequestWithContext(ctx, method, "http://yanakipre-fakeurl", nil)
			require.NoError(t, err)
			policy := StraightforwardNetworkRetry()
			policyDesicion := policy(connRstErr, r)
			require.True(t, IsPermanentError(policyDesicion)) // we do not retry this request
		})
	}

	for _, method := range []string{"GET", "HEAD", "OPTIONS"} {
		t.Run(fmt.Sprintf("%s retries unexpected EOF", method), func(t *testing.T) {
			r, err := http.NewRequestWithContext(ctx, method, "http://yanakipre-fakeurl", nil)
			require.NoError(t, err)

			policy := StraightforwardNetworkRetry()
			policyDecision := policy(io.ErrUnexpectedEOF, r)
			require.Error(t, policyDecision)
			require.False(t, IsPermanentError(policyDecision)) // we retry this request
		})
	}
	for _, method := range []string{"POST", "DELETE", "PUT"} {
		t.Run(fmt.Sprintf("%s doesn't retry unexpected EOF", method), func(t *testing.T) {
			r, err := http.NewRequestWithContext(ctx, method, "http://yanakipre-fakeurl", nil)
			require.NoError(t, err)

			policy := StraightforwardNetworkRetry()
			policyDecision := policy(io.ErrUnexpectedEOF, r)
			require.True(t, IsPermanentError(policyDecision)) // we do not retry this request
		})
	}
}

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
