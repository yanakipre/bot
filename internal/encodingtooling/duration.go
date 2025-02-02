package encodingtooling

import (
	"encoding/json"
	"fmt"
	"time"

	"gopkg.in/yaml.v2"
)

var (
	_ yaml.Marshaler   = &Duration{}
	_ yaml.Unmarshaler = &Duration{}
	_ json.Marshaler   = &Duration{}
	_ json.Unmarshaler = &Duration{}
)

// Duration wraps time.Duration for JSON formatting
// See for more https://github.com/golang/go/issues/10275
type Duration struct {
	// swagger:ignore
	Duration time.Duration
}

func NewDuration(dur time.Duration) Duration { return Duration{Duration: dur} }
func (d Duration) Ptr() *Duration            { return &d }
func (d Duration) String() string            { return d.Duration.String() }

func (d Duration) MarshalYAML() (any, error) {
	return d.String(), nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprint(`"`, d.String(), `"`)), nil
}

// UnmarshalYAML unmarshals Duration from string
func (d *Duration) UnmarshalYAML(unmarshal func(any) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	v, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	d.Duration = v
	return nil
}

// UnmarshalJSON unmarshals Duration from string
func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	v, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	d.Duration = v
	return nil
}
