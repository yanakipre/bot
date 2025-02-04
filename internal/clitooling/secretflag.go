package clitooling

import (
	"github.com/yanakipre/bot/internal/secret"
)

// SecretFlag represents command line flag with secret data
type SecretFlag struct {
	v secret.String
}

func (s *SecretFlag) String() string {
	return s.v.String()
}

func (s *SecretFlag) Get() *secret.String {
	return &(s.v)
}

func (s *SecretFlag) Set(s2 string) error {
	s.v = secret.NewString(s2)
	return nil
}

func (s *SecretFlag) Type() string {
	return "secret string"
}
