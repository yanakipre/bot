//nolint:lll
package staticconfig

import (
	"errors"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/openaiclient/httpopenaiclient"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/storage/postgres"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/controllers/controllerv1"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/transport/bottransport"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/transport/bottransportv2"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/transport/usertransport"

	"github.com/yanakipre/bot/internal/logger"
)

type Config struct {
	Ctlv1             controllerv1.Config
	PostgresRW        postgres.Config         `yaml:"postgres_rw"`
	OpenAI            httpopenaiclient.Config `yaml:"openai"`
	Logging           logger.Config           `yaml:"logging"`
	TelegramTransport bottransport.Config     `yaml:"telegram_transport"`
	TelegramV2        bottransportv2.Config   `yaml:"telegram_v2"`
	UserTransport     usertransport.Config    `yaml:"user_transport"`
}

func DefaultConfig() Config {
	return Config{
		Ctlv1:         controllerv1.DefaultConfig(),
		OpenAI:        httpopenaiclient.DefaultConfig(),
		Logging:       logger.DefaultConfig(),
		TelegramV2:    bottransportv2.DefaultConfig(),
		UserTransport: usertransport.DefaultConfig(),
	}
}

// All sub validations should go here.
func (c *Config) Validate() error {
	return errors.Join(
		c.TelegramV2.Validate(),
		c.UserTransport.Validate(),
	)
}
