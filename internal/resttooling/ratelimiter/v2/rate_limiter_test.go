package ratelimiter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yanakipre/bot/internal/rate"
	"github.com/yanakipre/bot/internal/testtooling"
)

func TestManager_parseConfig(t *testing.T) {
	testtooling.SetNewGlobalLoggerQuietly()

	tests := map[string]struct {
		cfg     Config
		manager *Manager
		check   func(t *testing.T, err error, m *Manager)
	}{
		"orders patterns in the order of precedence": {
			cfg: Config{
				Paths: map[string][]rate.WindowConfig{
					"/foo/{all...}": {},
					"/foo/{id}":     {},
					"/foo/bar":      {},
					"GET /foo/bar":  {},
				},
			},
			check: func(t *testing.T, err error, m *Manager) {
				require.Nil(t, err)
				require.Equal(
					t, []*pattern{
						{
							pattern:  "GET /foo/bar",
							method:   "GET",
							segments: []segment{{s: "foo"}, {s: "bar"}},
						},
						{
							pattern:  "/foo/bar",
							segments: []segment{{s: "foo"}, {s: "bar"}},
						},
						{
							pattern:  "/foo/{id}",
							segments: []segment{{s: "foo"}, {s: "id", identifier: true}},
						},
						{
							pattern:  "/foo/{all...}",
							segments: []segment{{s: "foo"}, {s: "all", identifier: true, multi: true}},
						},
					}, m.patterns,
				)
			},
		},
		"allows configuration from the RFC": {
			cfg: Config{
				Paths: map[string][]rate.WindowConfig{
					"/{all...}":      {},
					"POST /{all...}": {},
					"POST /pageservers/{pageserver_id}/migrate_projects": {},
				},
			},
			check: func(t *testing.T, err error, m *Manager) {
				require.Nil(t, err)
				require.Equal(
					t, []*pattern{
						{
							method:  "POST",
							pattern: "POST /pageservers/{pageserver_id}/migrate_projects",
							segments: []segment{
								{s: "pageservers"}, {s: "pageserver_id", identifier: true}, {s: "migrate_projects"},
							},
						},
						{
							method:   "POST",
							pattern:  "POST /{all...}",
							segments: []segment{{s: "all", identifier: true, multi: true}},
						},
						{
							pattern:  "/{all...}",
							segments: []segment{{s: "all", identifier: true, multi: true}},
						},
					}, m.patterns,
				)
			},
		},
		"allows reloading the manager state": {
			cfg: Config{
				Paths: map[string][]rate.WindowConfig{
					"/foo/{id}":     {},
					"/foo/{all...}": {},
				},
			},
			manager: &Manager{
				patterns: []*pattern{
					{
						pattern:  "/foo/{all...}",
						segments: []segment{{s: "foo"}, {s: "all", identifier: true, multi: true}},
					},
				},
			},
			check: func(t *testing.T, err error, m *Manager) {
				require.Nil(t, err)
				require.Equal(
					t, []*pattern{
						{
							pattern:  "/foo/{id}",
							segments: []segment{{s: "foo"}, {s: "id", identifier: true}},
						},
						{
							pattern:  "/foo/{all...}",
							segments: []segment{{s: "foo"}, {s: "all", identifier: true, multi: true}},
						},
					}, m.patterns,
				)
			},
		},
		"returns an error if the pattern is invalid": {
			cfg: Config{
				Paths: map[string][]rate.WindowConfig{
					"/foo/{id": {},
				},
			},
			check: func(t *testing.T, err error, m *Manager) {
				require.Error(t, err)
				require.ErrorContains(t, err, "failed to parse the path pattern")
			},
		},
		"returns an error if the patterns conflict": {
			cfg: Config{
				Paths: map[string][]rate.WindowConfig{
					"/foo/{bar}": {},
					"/foo/{baz}": {},
				},
			},
			check: func(t *testing.T, err error, m *Manager) {
				require.Error(t, err)
				require.ErrorContains(t, err, "conflicting path patterns")
			},
		},
	}

	ctx := context.Background()
	for name, tt := range tests {
		t.Run(
			name, func(t *testing.T) {
				if tt.manager == nil {
					tt.manager = &Manager{}
				}

				err := tt.manager.parseConfig(ctx, tt.cfg)
				tt.check(t, err, tt.manager)
			},
		)
	}
}
