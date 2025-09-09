package controllerv1

import (
	"time"

	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/openaiclient/httpopenaiclient"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/storage/postgres"

	"github.com/jellydator/ttlcache/v3"
)

type Ctl struct {
	// maps msgID to the ExplainedMessage
	explainedMessagesCache *ttlcache.Cache[int, ExplainedMessage]
	cfg                    Config
	openai                 *httpopenaiclient.Client
	storageRW              *postgres.Storage
}

func New(
	cfg Config,
	openai *httpopenaiclient.Client,
	storageRW *postgres.Storage,
) (*Ctl, error) {
	cache := ttlcache.New[int, ExplainedMessage](
		ttlcache.WithTTL[int, ExplainedMessage](30*time.Minute),
		// 200 MiB capacity
		ttlcache.WithCapacity[int, ExplainedMessage](1024*1024*200),
	)

	return &Ctl{
		explainedMessagesCache: cache,
		cfg:                    cfg,
		openai:                 openai,
		storageRW:              storageRW,
	}, nil
}

func (c *Ctl) Ready() error {
	go c.explainedMessagesCache.Start()
	return nil
}
