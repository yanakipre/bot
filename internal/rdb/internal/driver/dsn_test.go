package driver

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseOpts(t *testing.T) {
	tests := []struct {
		in       string
		expected values
		valid    bool
	}{
		{"dbname=hello;userid=goodbye", values{"dbname": "hello", "userid": "goodbye"}, true},
		{"dbname=hello user=goodbye", values{"dbname": "hello", "user": "goodbye"}, true},
		{"dbname=hello user=goodbye  ", values{"dbname": "hello", "user": "goodbye"}, true},
		{"dbname = hello user=goodbye", values{"dbname": "hello", "user": "goodbye"}, true},
		{"dbname = hello; user=goodbye", values{"dbname": "hello", "user": "goodbye"}, true},
		{"dbname=hello user =goodbye", values{"dbname": "hello", "user": "goodbye"}, true},
		{"dbname=hello user= goodbye", values{"dbname": "hello", "user": "goodbye"}, true},
		{
			"host=localhost password='correct horse battery staple'",
			values{"host": "localhost", "password": "correct horse battery staple"},
			true,
		},
		{"dbname=hello user=''", values{"dbname": "hello", "user": ""}, true},
		{"user='' dbname=hello", values{"dbname": "hello", "user": ""}, true},
		// The last option value is an empty string if there's no non-whitespace after its =
		{"dbname=hello user=   ", values{"dbname": "hello", "user": ""}, true},

		// The parser ignores spaces after = and interprets the next set of non-whitespace
		// characters as the value.
		{"user= password=foo", values{"user": "password=foo"}, true},

		// Backslash escapes next char
		{`user=a\ \'\\b`, values{"user": `a '\b`}, true},
		{`user='a \'b'`, values{"user": `a 'b`}, true},

		// Incomplete escape
		{`user=x\`, values{}, false},

		// No '=' after the key
		{"postgre://marko@internet", values{}, false},
		{"dbname user=goodbye", values{}, false},
		{"user=foo blah", values{}, false},
		{"user=foo blah   ", values{}, false},

		// Unterminated quoted value
		{"dbname=hello user='unterminated", values{}, false},
	}

	for _, test := range tests {
		t.Run(test.in, func(t *testing.T) {
			o := make(values)
			err := parseOpts(test.in, o)
			switch {
			case err != nil && test.valid:
				t.Errorf("%q got unexpected error: %s", test.in, err)
			case err == nil && test.valid && !reflect.DeepEqual(test.expected, o):
				t.Errorf("%q got: %#v want: %#v", test.in, o, test.expected)
			case err == nil && !test.valid:
				t.Errorf("%q expected an error", test.in)
			}
		})
	}
}

func TestParseDSN(t *testing.T) {
	cases := []struct {
		name       string
		dsn        string
		expectHost string
		expectDb   string
		expectUser string
	}{
		{"postgres uri", "postgresql://", "", "", ""},
		{"postgres uri", "postgresql://localhost", "localhost", "", ""},
		{"postgres uri", "postgresql://localhost:5433", "localhost:5433", "", ""},
		{"postgres uri", "postgresql://localhost/mydb", "localhost", "mydb", ""},
		{"postgres uri", "postgresql://user@localhost", "localhost", "", "user"},
		{"postgres uri", "postgresql://user:secret@localhost", "localhost", "", "user"},
		{
			"postgres uri",
			"postgresql://other@localhost/otherdb?connect_timeout=10&application_name=myapp",
			"localhost",
			"otherdb",
			"other",
		},
		{
			"postgres uri",
			"postgresql://host1:123,host2:456/somedb?target_session_attrs=any&application_name=myapp",
			"host1:123,host2:456",
			"somedb",
			"",
		},
		{
			"postgres query",
			"postgresql://host1:123,host2:456/somedb?target_session_attrs=any&application_name=myapp",
			"host1:123,host2:456",
			"somedb",
			"",
		},
		{
			"opts",
			"host=127.0.0.1 port=5436 database=some-service user=some-service-user sslmode=disable",
			"127.0.0.1:5436",
			"some-service",
			"some-service-user",
		},
	}

	for _, tcase := range cases {
		t.Run(tcase.name, func(t *testing.T) {
			h, db, u := parseDSN(tcase.dsn)
			assert.Equal(t, tcase.expectHost, h)
			assert.Equal(t, tcase.expectDb, db)
			assert.Equal(t, tcase.expectUser, u)
		})
	}
}
