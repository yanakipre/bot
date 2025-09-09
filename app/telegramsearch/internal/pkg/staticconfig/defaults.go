package staticconfig

import (
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/openaiclient/httpopenaiclient"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/storage/postgres"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/controllers/controllerv1"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/transport/usertransport"
	"github.com/yanakipre/bot/internal/logger"
)

func (c *Config) DefaultConfig() {
	c.Ctlv1 = controllerv1.DefaultConfig()
	c.OpenAI = httpopenaiclient.DefaultConfig()
	c.PostgresRW = postgres.Default()
	c.Logging = logger.DefaultConfig()
	c.UserTransport = usertransport.DefaultConfig()
}
