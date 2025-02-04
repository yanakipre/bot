package bottransport

import "github.com/yanakipre/bot/internal/secret"

type Config struct {
	Token    secret.String `yaml:"token"`
	Greeting string        `yaml:"greeting"`
}
