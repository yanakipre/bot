package ratelimiter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_newPattern(t *testing.T) {
	tests := map[string]struct {
		pattern string
		want    *pattern
		wantErr bool
	}{
		// Happy paths
		"No method specified": {
			pattern: "/foo",
			want: &pattern{
				segments: []segment{
					{s: "foo"},
				},
			},
		},
		"Simple GET": {
			pattern: "GET /foo",
			want: &pattern{
				method: "GET",
				segments: []segment{
					{s: "foo"},
				},
			},
		},
		"Simple GET with a trailing slash": {
			pattern: "GET /foo/",
			want: &pattern{
				method: "GET",
				segments: []segment{
					{s: "foo"},
				},
			},
		},
		"POST with multiple segments": {
			pattern: "POST /foo/bar/baz",
			want: &pattern{
				method: "POST",
				segments: []segment{
					{s: "foo"}, {s: "bar"}, {s: "baz"},
				},
			},
		},
		"PUT with an identifier": {
			pattern: "PUT /foo/{id}",
			want: &pattern{
				method: "PUT",
				segments: []segment{
					{s: "foo"}, {s: "id", identifier: true},
				},
			},
		},
		"PATCH with an identifier that matches across segments": {
			pattern: "PATCH /foo/{id}/{all...}",
			want: &pattern{
				method: "PATCH",
				segments: []segment{
					{s: "foo"}, {s: "id", identifier: true}, {s: "all", identifier: true, multi: true},
				},
			},
		},
		// Less happy paths
		"Empty pattern": {
			pattern: "",
			wantErr: true,
		},
		"Path missing a slash": {
			pattern: "GET foo",
			wantErr: true,
		},
		"Invalid method": {
			pattern: "FOO /foo",
			wantErr: true,
		},
		"Bad identifier format - prefix": {
			pattern: "GET /foo/i{d}",
			wantErr: true,
		},
		"Bad identifier format - suffix": {
			pattern: "GET /foo/{i}d",
			wantErr: true,
		},
		"Multi identifier not at the end": {
			pattern: "GET /foo/{all...}/bar",
			wantErr: true,
		},
		"Empty identifier name": {
			pattern: "GET /foo/{}",
			wantErr: true,
		},
		"Duplicate identifier name": {
			pattern: "GET /foo/{id}/{id}",
			wantErr: true,
		},
	}
	for name, tt := range tests {
		t.Run(
			name, func(t *testing.T) {
				got, err := newPattern(tt.pattern)
				if (err != nil) != tt.wantErr {
					t.Errorf("NewPattern() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if tt.want != nil {
					tt.want.pattern = tt.pattern
				}

				require.Equal(t, tt.want, got)
			},
		)
	}
}

func Test_patternMatch(t *testing.T) {
	tests := map[string]struct {
		pattern string
		method  string
		path    string
		want    bool
	}{
		"method mismatch": {
			pattern: "GET /foo",
			method:  "POST",
			path:    "/foo",
			want:    false,
		},
		"simple match": {
			pattern: "GET /foo",
			method:  "GET",
			path:    "/foo",
			want:    true,
		},
		"simple mismatch": {
			pattern: "GET /foo",
			method:  "GET",
			path:    "/bar",
			want:    false,
		},
		"pattern longer than path": {
			pattern: "GET /foo/bar",
			method:  "GET",
			path:    "/foo",
			want:    false,
		},
		"pattern longer than path with multi": {
			pattern: "GET /foo/bar/{all...}",
			method:  "GET",
			path:    "/foo/",
			want:    false,
		},
		"path longer than pattern, no multi": {
			pattern: "GET /foo",
			method:  "GET",
			path:    "/foo/bar",
			want:    false,
		},
		"match with an identifier in the middle": {
			pattern: "POST /foo/{id}/bar",
			method:  "POST",
			path:    "/foo/123/bar",
			want:    true,
		},
		"multi matches everything": {
			pattern: "GET /foo/{all...}",
			method:  "GET",
			path:    "/foo/bar/baz",
			want:    true,
		},
	}
	for name, tt := range tests {
		t.Run(
			name, func(t *testing.T) {
				p, err := newPattern(tt.pattern)
				if err != nil {
					panic(err)
				}
				path := tt.path[1:]
				segments := strings.Split(path, "/")
				require.Equal(t, tt.want, p.match(tt.method, segments))
			},
		)
	}
}
