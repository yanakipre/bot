package ctl

import (
	"context"
	"fmt"

	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/openaiclient/httpopenaiclient"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/storage/postgres"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/controllers/controllerv1"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/staticconfig"
)

func Init(ctx context.Context, staticConfig *staticconfig.Config) (*controllerv1.Ctl, error) {
	storageRW := postgres.New(staticConfig.PostgresRW)
	err := storageRW.Ready(ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating storage: %w", err)
	}

	openai := httpopenaiclient.NewClient(staticConfig.OpenAI)

	ctl, err := controllerv1.New(staticConfig.Ctlv1, openai, storageRW)
	if err != nil {
		return nil, fmt.Errorf("error creating controller: %w", err)
	}
	return ctl, nil
}
