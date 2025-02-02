package config

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/yanakipe/bot/internal/encodingtooling"
	"github.com/yanakipe/bot/internal/secret"
)

type EmptyMap struct {
	Field map[string]struct{} `yaml:"metric_id_to_group_by"`
}

func (c *EmptyMap) DefaultConfig() {
}

func (c *EmptyMap) Validate() error {
	return nil
}

type DurationTestSingleField struct {
	LoadDuration encodingtooling.Duration `yaml:"timeout"`
}

func (t *DurationTestSingleField) DefaultConfig() {
}

func (t *DurationTestSingleField) Validate() error {
	return nil
}

type NestedCfg struct {
	BoolField bool `yaml:"bool_field"`
}

type TestCfg struct {
	StringField string
	Nested      NestedCfg     `yaml:"nested"`
	SecretField secret.String `yaml:"secret"`
}

func (t *TestCfg) DefaultConfig() {
	t.StringField = "test-value"
}

func (t *TestCfg) Validate() error {
	return nil
}

type ListCfg struct {
	Val string `yaml:"val"`
}

type TestListCfg struct {
	Values []ListCfg `yaml:"values"`
}

func (t *TestListCfg) DefaultConfig() {
}

func (t *TestListCfg) Validate() error {
	return nil
}

type DurationTest struct {
	LoadDuration encodingtooling.Duration `yaml:"timeout"`
	SomeField    string                   `yaml:"some_field"`
}

func (t *DurationTest) DefaultConfig() {
}

func (t *DurationTest) Validate() error {
	return nil
}

type ErrorTest struct {
	SomeField string `yaml:"some_field"`
}

func (t *ErrorTest) DefaultConfig() {
}

func (t *ErrorTest) Validate() error {
	return errors.New("validation error")
}

func TestLoad(t *testing.T) {
	ctx := context.Background()

	type args struct {
		appName     string
		unmarshalTo Config
		filename    string
	}
	tests := []struct {
		name        string
		args        args
		prepare     func()
		teardown    func()
		expect      Config
		expectError error
	}{
		{
			// confita cannot guess how to unmarshal struct with single encodingtooling.Duration
			// field defined.
			name: "single duration field is not unmarshalled",
			args: args{
				unmarshalTo: &DurationTestSingleField{},
				filename:    "test_duration.yaml",
			},
			expect: &DurationTestSingleField{
				LoadDuration: encodingtooling.Duration{}, // so the value is empty
			},
		},
		{
			name: "empty map on next line",
			args: args{
				unmarshalTo: &EmptyMap{},
				filename:    "test_empty_map.yaml",
			},
			expect: &EmptyMap{
				Field: map[string]struct{}{},
			},
		},
		{
			name: "duration",
			args: args{
				unmarshalTo: &DurationTest{},
				filename:    "test_duration.yaml",
			},
			expect: &DurationTest{
				LoadDuration: encodingtooling.Duration{Duration: time.Hour + 5*time.Second},
				SomeField:    "1",
			},
		},
		{
			name: "list values",
			args: args{
				unmarshalTo: &TestListCfg{},
				filename:    "test_list.yaml",
			},
			expect: &TestListCfg{
				Values: []ListCfg{
					{Val: "1"},
					{Val: "2"},
				},
			},
		},
		{
			name: "secrets from env are not loaded automatically",
			args: args{
				unmarshalTo: &TestCfg{},
				filename:    "",
			},
			expect: &TestCfg{
				StringField: "test-value",
				SecretField: secret.NewString(""), // the value is EMPTY
			},
			prepare: func() {
				if err := os.Setenv("secret", "hello"); err != nil {
					panic(err)
				}
			},
			teardown: func() {
				if err := os.Unsetenv("secret"); err != nil {
					panic(err)
				}
			},
		},
		{
			name: "secrets from config",
			args: args{
				unmarshalTo: &TestCfg{},
				filename:    "test_secret.yaml",
			},
			expect: &TestCfg{
				StringField: "test-value",
				SecretField: secret.NewString("hello"),
			},
		},
		{
			name: "defaults overwritten with environment",
			args: args{
				unmarshalTo: &TestCfg{},
				filename:    "",
			},
			expect: &TestCfg{
				StringField: "test-value",
				Nested: NestedCfg{
					BoolField: true,
				},
			},
			prepare: func() {
				if err := os.Setenv("bool_field", "true"); err != nil {
					panic(err)
				}
			},
			teardown: func() {
				if err := os.Unsetenv("bool_field"); err != nil {
					panic(err)
				}
			},
		},
		{
			name: "defaults overwritten with file",
			args: args{
				unmarshalTo: &TestCfg{},
				filename:    "test_config.yaml",
			},
			expect: &TestCfg{
				StringField: "test-value",
				Nested: NestedCfg{
					BoolField: true,
				},
			},
		},
		{
			name: "defaults applied",
			args: args{
				unmarshalTo: &TestCfg{},
				filename:    "",
			},
			expect: &TestCfg{
				StringField: "test-value",
			},
		},
		{
			name: "error in validation",
			args: args{
				unmarshalTo: &ErrorTest{},
				filename:    "",
			},
			expectError: errors.New("validation error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.prepare != nil {
				tt.prepare()
			}
			cfg := tt.args.unmarshalTo
			err := Load(ctx, tt.args.appName, cfg, tt.args.filename)
			if tt.expectError != nil {
				require.Equal(t, tt.expectError, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expect, cfg)
			if tt.prepare != nil {
				tt.teardown()
			}
		})
	}
}
