package secret

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
)

const maskedString = "xxx"

type String = Value[string]

func NewString(in string) String { return NewValue(in) }

// String protects value from accidental exposure. It is a struct with private attribute because
// that way the value is hidden from reflection.
type Value[T any] struct {
	value T
}

// Scan implements the sql Scanner interface.
func (s *Value[T]) Scan(src any) error {
	v, ok := src.(T)
	if !ok {
		return errors.New("type assertion for secret failed")
	}
	s.value = v
	return nil
}

// Value implements the sql driver Valuer interface.
func (s Value[T]) Value() (driver.Value, error) {
	return s.value, nil
}

// NewValue[T] constructs secret value of type string
func NewValue[T any](v T) Value[T] {
	return Value[T]{value: v}
}

// Value[T] returns masked string
func (Value[T]) String() string {
	return maskedString
}

func (Value[T]) Format(s fmt.State, verb rune) { _, _ = fmt.Fprint(s, maskedString) }

func (s Value[T]) Ptr() *Value[T] { return &s }

// Unmask returns real value
func (s Value[T]) Unmask() T {
	return s.value
}

// UnmarshalJSON allows parsing secret from JSON
func (s *Value[T]) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &s.value)
}

// MarshalJSON hides secret as pure string
func (s Value[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalYAML allows parsing secret from YAML
func (s *Value[T]) UnmarshalYAML(unmarshal func(any) error) error {
	return unmarshal(&s.value)
}

// MarshalYAML hides secret as pure string
func (s Value[T]) MarshalYAML() (any, error) {
	return s.String(), nil
}

// FromEnv allows loading secret strings from environment variables.
func (s *Value[T]) FromEnv(name string) bool {
	v, ok := os.LookupEnv(name)
	if ok {
		switch r := any(s.value).(type) {
		case string:
			if v == "" {
				// skip empty strings to ease local development,
				// when someone is using .env file with empty secrets
				return false
			}
			switch val := any(v).(type) {
			case T:
				s.value = val
				return true
			}
		case bool:
			var unwrapped any
			if v == "true" || v == "True" {
				unwrapped = true
			} else if v == "false" || v == "False" {
				unwrapped = false
			}
			switch val := unwrapped.(type) {
			case T:
				s.value = val
				return true
			default:
				return false
			}
		case int:
			var unwrapped any
			if num, err := strconv.Atoi(v); err == nil {
				unwrapped = num
			}
			switch val := unwrapped.(type) {
			case T:
				s.value = val
				return true
			default:
				return false
			}
		case json.Unmarshaler:
			if err := r.UnmarshalJSON([]byte(v)); err == nil {
				return true
			}
		}
	}
	return false
}
