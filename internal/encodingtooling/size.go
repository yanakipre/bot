package encodingtooling

import (
	"encoding/json"

	"github.com/docker/go-units"
	"gopkg.in/yaml.v2"
)

var (
	_ yaml.Marshaler   = &Size{}
	_ yaml.Unmarshaler = &Size{}
	_ json.Marshaler   = &Size{}
	_ json.Unmarshaler = &Size{}
)

// Size wraps int64 for JSON formatting
type Size struct {
	// swagger:ignore
	Size int64 // Bytes
}

func NewSize(size int64) Size { return Size{Size: size} }

func (d Size) String() string {
	return units.BytesSize(float64(d.Size))
}

// MarshalYAML marshals Size to string
func (d Size) MarshalYAML() (any, error) {
	return d.String(), nil
}

// UnmarshalYAML unmarshals Size from string
func (d *Size) UnmarshalYAML(unmarshal func(any) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	v, err := units.RAMInBytes(s)
	if err != nil {
		return err
	}

	d.Size = v
	return nil
}

// MarshalJSON marshals Size to string
func (d Size) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.String() + `"`), nil
}

// UnmarshalJSON unmarshals Size from string
func (d *Size) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	v, err := units.RAMInBytes(s)
	if err != nil {
		return err
	}

	d.Size = v
	return nil
}
