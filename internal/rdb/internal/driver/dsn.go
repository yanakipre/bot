package driver

import (
	"errors"
	"fmt"
	"net"
	nurl "net/url"
	"unicode"

	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/clouderr"
)

func parseDSN(dsn string) (host string, db string, user string) {
	var v values
	var err error
	if v, err = parseURL(dsn); err != nil {
		v = values{}
		if err = parseOpts(dsn, v); err != nil {
			v = values{}
			values, err := nurl.ParseQuery(dsn)
			if err == nil {
				for k, val := range values {
					v[k] = val[0]
				}
			}
		}
	}

	if values := v["host"]; values != "" {
		host = values
	} else if values := v["server"]; values != "" {
		host = values
	}
	if values := v["port"]; values != "" {
		host += ":" + values
	}
	if values := v["database"]; values != "" {
		db = values
	}
	if values := v["user"]; values != "" {
		user = values
	} else if values := v["user id"]; values != "" {
		user = values
	}
	return
}

// scanner implements a tokenizer for libpq-style option strings.
type scanner struct {
	s []rune
	i int
}

// newScanner returns a new scanner initialized with the option string s.
func newScanner(s string) *scanner {
	return &scanner{[]rune(s), 0}
}

// Next returns the next rune.
// It returns 0, false if the end of the text has been reached.
func (s *scanner) Next() (rune, bool) {
	if s.i >= len(s.s) {
		return 0, false
	}
	r := s.s[s.i]
	s.i++
	return r, true
}

// SkipSpaces returns the next non-whitespace rune.
// It returns 0, false if the end of the text has been reached.
func (s *scanner) SkipSpaces() (rune, bool) {
	r, ok := s.Next()
	for (unicode.IsSpace(r) /*|| r == ';'*/) && ok {
		r, ok = s.Next()
	}
	return r, ok
}

type values map[string]string

// parseOpts parses the options from name and adds them to the values.
//
// The parsing code is based on conninfo_parse from libpq's fe-connect.c
func parseOpts(name string, o values) error {
	s := newScanner(name)

	for {
		var (
			keyRunes, valRunes []rune
			r                  rune
			ok                 bool
		)

		if r, ok = s.SkipSpaces(); !ok {
			break
		}

		// Scan the key
		for !(unicode.IsSpace(r)) && r != '=' {
			keyRunes = append(keyRunes, r)
			if r, ok = s.Next(); !ok {
				break
			}
		}

		// Skip any whitespace if we're not at the = yet
		if r != '=' {
			r, ok = s.SkipSpaces()
		}

		// The current character should be =
		if r != '=' || !ok {
			return clouderr.WithFields(
				`missing "=" after key in connection info string"`,
				zap.String("key", string(keyRunes)),
			)
		}

		// Skip any whitespace after the =
		if r, ok = s.SkipSpaces(); !ok {
			// If we reach the end here, the last value is just an empty string as per libpq.
			o[string(keyRunes)] = ""
			break
		}

		if r != '\'' {
			for !(unicode.IsSpace(r) || r == ';') {
				if r == '\\' {
					if r, ok = s.Next(); !ok {
						return fmt.Errorf(`missing character after backslash`)
					}
				}
				valRunes = append(valRunes, r)

				if r, ok = s.Next(); !ok {
					break
				}
			}
		} else {
		quote:
			for {
				if r, ok = s.Next(); !ok {
					return fmt.Errorf(`unterminated quoted string literal in connection string`)
				}
				switch r {
				case '\'':
					break quote
				case '\\':
					r, _ = s.Next()
					fallthrough
				default:
					valRunes = append(valRunes, r)
				}
			}
		}

		o[string(keyRunes)] = string(valRunes)
	}

	return nil
}

func parseURL(url string) (values, error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "" {
		return nil, errors.New("wrong schema")
	}
	kvs := values{}
	accrue := func(k, v string) {
		if v != "" {
			kvs[k] = v
		}
	}

	if u.User != nil {
		v := u.User.Username()
		accrue("user", v)

		v, _ = u.User.Password()
		accrue("password", v)
	}

	if host, port, err := net.SplitHostPort(u.Host); err != nil {
		accrue("host", u.Host)
	} else {
		accrue("host", host)
		accrue("port", port)
	}

	if u.Path != "" {
		accrue("database", u.Path[1:])
	}

	q := u.Query()
	for k := range q {
		accrue(k, q.Get(k))
	}
	return kvs, nil
}
