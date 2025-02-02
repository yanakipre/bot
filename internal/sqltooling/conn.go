package sqltooling

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/yanakipe/bot/internal/secret"
)

type ConnStringKV struct {
	Host     string
	Database string
	User     string
	Password secret.String
	Port     int
	SSLMode  string
}

const PostgresDefaultPort = 5432

func escapeForPostgres(v string) string {
	return strings.Replace(strings.Replace(
		v,
		"\\", "\\\\", 100),
		"'", "\\'", 100)
}

// https://www.postgresql.org/docs/14/libpq-connect.html#id-1.7.3.8.3.5
func (cs *ConnStringKV) KVConnectionString() secret.String {
	type kvP struct {
		k string
		v string
	}
	kvPairs := make([]kvP, 0, 7)
	kvPairs = append(kvPairs, kvP{"host", cs.Host})
	kvPairs = append(kvPairs, kvP{"database", cs.Database})
	kvPairs = append(kvPairs, kvP{"user", cs.User})
	kvPairs = append(kvPairs, kvP{"password", cs.Password.Unmask()})
	kvPairs = append(kvPairs, kvP{"port", fmt.Sprintf("%d", cs.Port)})
	if cs.SSLMode == "" {
		kvPairs = append(kvPairs, kvP{"sslmode", "disable"})
	} else {
		kvPairs = append(kvPairs, kvP{"sslmode", cs.SSLMode})
	}
	kvPairs = append(kvPairs, kvP{"client_encoding", "UTF8"})

	var resultSlice []string
	for _, pair := range kvPairs {
		resultSlice = append(resultSlice, fmt.Sprintf("%s='%s'", pair.k, escapeForPostgres(pair.v)))
	}
	return secret.NewString(strings.Join(resultSlice, " "))
}

// ConnectionURI as in Postgres docs:
// https://www.postgresql.org/docs/current/libpq-connect.html#id-1.7.3.8.3.6
type ConnectionURI struct {
	Host     string
	Database string
	User     string
	Password secret.String
	Port     int
	Options  string
}

func (c ConnectionURI) ConnectionURI() secret.String {
	var host string
	if c.Port == PostgresDefaultPort || c.Port == 0 {
		host = c.Host
	} else {
		host = fmt.Sprintf("%s:%d", c.Host, c.Port)
	}
	uri := url.URL{
		Scheme: "postgresql",
		Host:   host,
		Path:   c.Database,
	}
	if c.Password.Unmask() != "" {
		uri.User = url.UserPassword(c.User, c.Password.Unmask())
	} else {
		uri.User = url.User(c.User)
	}
	q := uri.Query()
	q.Set("sslmode", "require")
	if c.Options != "" {
		q.Set("options", c.Options)
	}
	uri.RawQuery = q.Encode()
	return secret.NewString(uri.String())
}
