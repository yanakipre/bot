package staticconfig

import (
	"github.com/yanakipe/bot/app/telegramsearch/internal/pkg/client/openaiclient/httpopenaiclient"
	"github.com/yanakipe/bot/app/telegramsearch/internal/pkg/client/storage/postgres"
	"github.com/yanakipe/bot/app/telegramsearch/internal/pkg/controllers/controllerv1"
	"github.com/yanakipe/bot/internal/logger"
)

func (c *Config) DefaultConfig() {
	c.Ctlv1 = controllerv1.DefaultConfig()
	c.OpenAI = httpopenaiclient.DefaultConfig()
	c.PostgresRW = postgres.Default()
	c.Logging = logger.DefaultConfig()
}
