package controllerv1

import (
	"context"
	"errors"
	"github.com/yanakipe/bot/app/telegramsearch/internal/pkg/client/openaiclient/openaimodels"
	"github.com/yanakipe/bot/app/telegramsearch/internal/pkg/client/storage/storagemodels"
	models "github.com/yanakipe/bot/app/telegramsearch/internal/pkg/controllers/controllerv1/controllerv1models"
	"sync"
	"time"

	"github.com/samber/lo"
	"github.com/sourcegraph/conc/pool"
	"github.com/yanakipe/bot/internal/logger"
	"go.uber.org/zap"
)

func (c *Ctl) TryEmbedding(ctx context.Context, req models.ReqTryEmbedding) (models.RespTryEmbedding, error) {
	queryResponse, err := c.openai.CreateEmbeddings(ctx, openaimodels.ReqCreateEmbeddings{
		Input: []string{req.Input},
	})
	if err != nil {
		return models.RespTryEmbedding{}, err
	}

	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()
	p := pool.New().WithContext(ctxWithCancel).WithCancelOnError()
	now := time.Now()
	minusDay := -time.Hour * 24
	ranges := []time.Time{
		//now.Add(365 * 10 * minusDay),
		now.Add(365 * 3 * minusDay),
		now.Add(365 * minusDay),
		now.Add(30 * minusDay),
		now.Add(7 * minusDay),
		now,
	}
	const maxResults = 30
	searchResults := make([]storagemodels.RespSimilaritySearch, 0, maxResults)
	mu := sync.Mutex{}
	for i := range ranges {
		if i == 0 {
			continue
		}
		l := ranges[i-1]
		r := ranges[i]
		p.Go(func(ctx context.Context) error {
			search, err := c.storageRW.FetchSimilaritySearch(ctx, storagemodels.ReqSimilaritySearch{
				CutThreshold: 0.6, // empirical value
				Embedding:    queryResponse.Embeddings[0].Embedding,
				Since:        l,
				UpTo:         r,
				Limit:        20,
			})
			if err != nil {
				switch {
				case errors.Is(err, context.Canceled):
					// if we got enough results - suppress, it's OK
					if len(searchResults) == maxResults {
						return nil
					}
					return err
				default:
					return err
				}
			}
			logger.Info(ctx, "got messages", zap.Int("count", len(search)), zap.Time("from", l), zap.Time("to", r))
			mu.Lock()
			defer mu.Unlock()
			for i := range search {
				if len(searchResults) == maxResults {
					cancel() // notify we got enough results
				}
				searchResults = append(searchResults, search[i])
			}
			return nil
		})
	}
	if err = p.Wait(); err != nil {
		return models.RespTryEmbedding{}, err
	}

	return models.RespTryEmbedding{
		Result: lo.Map(searchResults, func(item storagemodels.RespSimilaritySearch, _ int) models.EmbeddingResponse {
			return models.EmbeddingResponse{
				Text: item.Message,
			}
		}),
	}, nil
}
