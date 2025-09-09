package controllerv1

import (
	"context"
	"encoding/json"

	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/openaiclient/openaimodels"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/storage/storagemodels"
	models "github.com/yanakipre/bot/app/telegramsearch/internal/pkg/controllers/controllerv1/controllerv1models"

	"github.com/sourcegraph/conc/pool"
	"github.com/yanakipre/bot/internal/logger"
)

func (c *Ctl) GenerateEmbeddings(ctx context.Context, req models.ReqGenerateEmbeddings) (models.RespGenerateEmbeddings, error) {
	lg := logger.FromContext(ctx)
	for {
		threadsToGenerateFrom, err := c.storageRW.FetchChatThreadToGenerateEmbedding(ctx, storagemodels.ReqFetchChatThreadToGenerateEmbedding{})
		if err != nil {
			return models.RespGenerateEmbeddings{}, err
		}
		if len(threadsToGenerateFrom.Threads) == 0 {
			break // no more
		}
		p := pool.New().WithMaxGoroutines(100).WithContext(ctx)
		for i := range threadsToGenerateFrom.Threads {
			proccess := threadsToGenerateFrom.Threads[i]
			p.Go(func(ctx context.Context) error {
				var t thread
				err := json.Unmarshal(proccess.Body, &t)
				if err != nil {
					return err
				}
				if len(t) < 2 {
					return nil // we don't want threads without answers
				}
				queryResponse, err := c.openai.CreateEmbeddings(ctx, openaimodels.ReqCreateEmbeddings{
					Input: []string{t.ForEmbedding()},
				})
				msg, err := t.ForShowingToTheUser(proccess.ChatID)
				if err != nil {
					return err
				}
				if len(queryResponse.Embeddings) == 0 {
					lg.Warn("skipped, no embeddings")
					return nil
				}
				_, err = c.storageRW.UpsertEmbedding(ctx, storagemodels.ReqUpsertEmbedding{
					Embedding: queryResponse.Embeddings[0].Embedding,
					Message:   msg,
					ChatID:    proccess.ChatID,
					ThreadID:  proccess.ThreadID,
				})
				return err
			})
		}
		if err := p.Wait(); err != nil {
			return models.RespGenerateEmbeddings{}, err
		}
	}
	return models.RespGenerateEmbeddings{}, nil
}
