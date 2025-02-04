package resttooling

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/yanakipre/bot/internal/encodingtooling"
)

var (
	transportMtx   = &sync.Mutex{}
	transportCache = map[Config]http.RoundTripper{}
)

type Config struct {
	// ClientName is how the client will appear throughout the instrumentation:
	// in logs, metrics, etc.
	// Not expected to be configured by YAML configs.
	ClientName string `yaml:"client_name" json:"client_name,omitempty"`

	// RequestTimeout actually sets the "header_timeout" on the
	// http.Transport. We should receive the _header_ within the
	// scope of this timeout, but have a bit of additional
	// flexibility to read the response given the timeout.
	// DEPRECATED
	// use ResponseHeaderTimeout instead
	RequestTimeout encodingtooling.Duration `yaml:"timeout"                  json:"-"`
	// Client timeout is the high level timeout on the
	// http.Client. This is the total time that a
	// request. including consuming  the entire request and
	// handling retries.
	//
	// When not specified explicitly the timeout is the number of
	// retries (plus 1) times the request timeout.
	ClientTimeout *encodingtooling.Duration `yaml:"client_timeout,omitempty" json:"client_timeout,omitempty"`
	// Dial* options are used when creating the net.Dialer{}
	// object that provides the context
	DialTimeout   encodingtooling.Duration `yaml:"dial_timeout"             json:"dial_timeout"`
	DialKeepAlive encodingtooling.Duration `yaml:"dial_keep_alive"          json:"dial_keep_alive"`
	// These options have the same names and
	// semantics as their http.Transport counterparts.
	IdleConnTimeout       encodingtooling.Duration `yaml:"idle_conn_timeout"        json:"idle_conn_timeout"`
	TLSHandshakeTimeout   encodingtooling.Duration `yaml:"tls_handshake_timeout"    json:"tls_handshake_timeout"`
	ExpectContinueTimeout encodingtooling.Duration `yaml:"expect_continue_timeout"  json:"expect_continue_timeout"`
	ResponseHeaderTimeout encodingtooling.Duration `yaml:"response_header_timeout"  json:"response_header_timeout"`
	DisableKeepAlives     bool                     `yaml:"disable_keep_alives"      json:"disable_keep_alives"`
	MaxIdleConns          int                      `yaml:"max_idle_conns"           json:"max_idle_conns"`
	MaxIdleConnsPerHost   int                      `yaml:"max_idle_conns_per_host"  json:"max_idle_conns_per_host"`
	MaxConnsPerHost       int                      `yaml:"max_conns_per_host"       json:"max_conns_per_host"`
}

// DefaultTransportConfig provides default values for the
// http.Transport: the default values defined in this function are
// *exactly* the same as the ones provided by the http.Package.
//
// There are three structural differences that do not impact the
// semantics of these timeouts relative to the `http.DefaultTransport`:
//
//   - the "dial" settings have a different name than their stdlib
//     equivalents (but the same value)
//
//   - the types use our duration wrapper type to make them easier to
//     specify in configuration.
//
//   - MaxIdleConnsPerHost is explicitly set to the value that it
//     defaults to in the current net/http/transport code.
//
// There are two additional setting which is custom to yanakipre:
//
//   - MaxIdleConns, is usually unset (zero) which means unlimited. It
//     now has a limit, the exact value of which is arbitrary.
//
//   - ResponseHeaderTimeout, which is not set in the default
//
// The only provided value which defaults to its zero value and is not
// set by the default config is "DisableKeepAlives" which is false
// here and false in the default configuration.
func DefaultTransportConfig() Config {
	return Config{
		// These control the behavior of the net.Dialer used
		// when establishing the TCP socket. Dialing happens
		// concurrently with request processing, so a dial
		// option can take longer than a request, and the (idle)
		// connection pool can use these connections.
		//
		// The Defaults are 30s for both timeout and
		// keepalive. I've made both longer just to increase
		// the chance that an active connection doesn't get thrown
		// away.
		DialTimeout:   encodingtooling.NewDuration(time.Minute),
		DialKeepAlive: encodingtooling.NewDuration(5 * time.Minute),
		// MaxIdleConnections per host defaults to 2 in the
		// standard library, and prevents the pool
		// disproportionately using one host.
		MaxIdleConnsPerHost: http.DefaultMaxIdleConnsPerHost,
		// This controls the amount of time for the server to
		// respond to a request. It includes all parts of the
		// request (getting a connection, writing the request body to
		// the socket,) up to getting the headers back from
		// the server. The default, 0, provides no limit (in
		// effect) means that the requests other limits with
		// control the timeout.
		ResponseHeaderTimeout: encodingtooling.NewDuration(0),
		// This controls the number of idle connection in the
		// client's connection pool, and therefore the number
		// of resources used by the connection. The
		// MaxIdleConns defaults to 100, and we've doubled the
		// IdelConnTimeout.
		MaxIdleConns:    100,
		IdleConnTimeout: encodingtooling.NewDuration(3 * time.Minute),
		// These values are coppied directly from the standard
		// library's default transport. Hardcoded here mostly
		// to avoid needing to cross reference multiple
		// sources. If there were more, we should copy.
		TLSHandshakeTimeout:   encodingtooling.NewDuration(10 * time.Second),
		ExpectContinueTimeout: encodingtooling.NewDuration(time.Second),
	}
}

func (conf Config) Resolve() http.RoundTripper {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   conf.DialTimeout.Duration,
			KeepAlive: conf.DialKeepAlive.Duration,
		}).DialContext,
		// this mirrors the value in the default transport,
		// and we do not currently provide a configuration
		// option to mutate it. there seems to be little
		// reason.
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          conf.MaxIdleConns,
		MaxIdleConnsPerHost:   conf.MaxIdleConnsPerHost,
		MaxConnsPerHost:       conf.MaxConnsPerHost,
		IdleConnTimeout:       conf.IdleConnTimeout.Duration,
		TLSHandshakeTimeout:   conf.TLSHandshakeTimeout.Duration,
		ExpectContinueTimeout: conf.ExpectContinueTimeout.Duration,
		ResponseHeaderTimeout: conf.ResponseHeaderTimeout.Duration,
		DisableKeepAlives:     conf.DisableKeepAlives,
	}
}

func getHTTPRoundTripper(conf Config) http.RoundTripper {
	var zero Config
	if conf == zero {
		conf = DefaultTransportConfig()
	}
	transportMtx.Lock()
	defer transportMtx.Unlock()
	tr, ok := transportCache[conf]
	if !ok {
		tr = conf.Resolve()
		transportCache[conf] = tr
	}
	return tr
}

func roundTripperCacheSize() int {
	transportMtx.Lock()
	defer transportMtx.Unlock()
	return len(transportCache)
}
