package controllerv1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/openaiclient/openaimodels"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/storage/storagemodels"
	models "github.com/yanakipre/bot/app/telegramsearch/internal/pkg/controllers/controllerv1/controllerv1models"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/jellydator/ttlcache/v3"
	"github.com/samber/lo"
	"github.com/sourcegraph/conc/pool"
	"github.com/yanakipre/bot/internal/logger"
	"go.uber.org/zap"
)

func (c *Ctl) TryCompletion(ctx context.Context, req models.ReqTryCompletion) (models.RespTryCompletion, error) {
	logger.Info(ctx, "user asked for completion", zap.String("q", req.Query))

	queryResponse, err := c.openai.CreateEmbeddings(ctx, openaimodels.ReqCreateEmbeddings{
		Input: []string{req.Query},
	})
	if err != nil {
		return models.RespTryCompletion{}, fmt.Errorf("create embeddings: %w", err)
	}
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()
	p := pool.New().WithContext(ctxWithCancel).WithCancelOnError()
	now := time.Now()
	minusDay := -time.Hour * 24
	ranges := []time.Time{
		now.Add(365 * 10 * minusDay),
		now.Add(365 * 3 * minusDay),
		now.Add(365 * 2 * minusDay),
		now.Add(365 * minusDay),
		now.Add(180 * minusDay),
		now.Add(30 * minusDay),
		now.Add(7 * minusDay),
		now,
	}
	const maxResults = 20
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
				CutThreshold: 0.5, // empirical value
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
				searchResults = append(searchResults, search[i])
				if len(searchResults) == maxResults {
					cancel() // notify we got enough results
				}
			}
			return nil
		})
	}
	if err = p.Wait(); err != nil {
		if !errors.Is(err, context.Canceled) {
			// ignore those errors
			return models.RespTryCompletion{}, err
		}
	}
	if len(searchResults) == 0 {
		return models.RespTryCompletion{
			Response:          c.cfg.NoResultsAnswer,
			UsedConversations: searchResults,
		}, nil
	}
	slices.SortFunc(searchResults, func(a, b storagemodels.RespSimilaritySearch) int {
		return int(a.MostRecentMessageAt.Sub(b.MostRecentMessageAt).Nanoseconds())
	})
	slices.Reverse(searchResults)
	completion, err := c.openai.CreateChatCompletion(ctx, openaimodels.ReqCreateChatCompletion{
		Input: req.Query,
		Conversations: lo.Map(searchResults, func(item storagemodels.RespSimilaritySearch, _ int) string {
			logger.Info(ctx, "used for response", zap.String("thread", item.Message))
			return item.Message
		}),
	})
	onlyStaleResponses := true
	notStaleAfter := now.Add(-1 * c.cfg.StaleThreshold.Duration)
	for i := range searchResults {
		if searchResults[i].MostRecentMessageAt.After(notStaleAfter) {
			onlyStaleResponses = false
			break
		}
	}
	userResponse := completion.Response
	if onlyStaleResponses {
		userResponse = completion.Response + "\n" + fmt.Sprintf(c.cfg.StaleResponsesText, notStaleAfter.Format(time.DateOnly))
	} else if len(searchResults) < 3 {
		userResponse = completion.Response + "\n" + fmt.Sprintf(c.cfg.FreshResponsesText, len(searchResults))
	}
	if err != nil {
		return models.RespTryCompletion{}, fmt.Errorf("failed to create completion: %w", err)
	}

	// update cache
	if err := c.saveCacheItem(req.SenderID, searchResults); err != nil {
		logger.Error(ctx, "failed to save cache item", zap.Error(err))
	}

	return models.RespTryCompletion{
		Response:          userResponse,
		UsedConversations: searchResults,
	}, nil
}

func (c *Ctl) saveCacheItem(senderID int, conversations []storagemodels.RespSimilaritySearch) error {
	cacheItem := ExplainedMessage{
		Sources: make([]SourcedMessage, 0, len(conversations)),
	}

	limit := 15
	if len(conversations) < limit {
		limit = len(conversations)
	}
	logger.Warn(context.Background(), "saving cache item", zap.Int("count", limit), zap.Int("len", len(conversations)))

	for _, conv := range conversations[:limit] {
		// to get the first letters from the Message
		// extract at least 40 symbols,
		// at max 200 symbols and cut everything in between by a dot ".".
		// if the message is shorter than 40 symbols, take the whole message.
		var s []serializedChatMessage
		err := json.Unmarshal([]byte(conv.ConversationStarter), &s)
		if err != nil {
			return fmt.Errorf("failed to unmarshal conversation starter: %w", err)
		}
		firstLetters := s[0].getText()
		if len(firstLetters) > 60 {
			lastDot := strings.Index(firstLetters, ".")
			if lastDot != -1 {
				firstLetters = firstLetters[:lastDot+1]
			}
		}
		if len(firstLetters) > 120 {
			firstLetters = firstLetters[:100]
		}
		cacheItem.Sources = append(cacheItem.Sources, SourcedMessage{
			URL:                 fmt.Sprintf("https://t.me/%s/%d", conv.TelegramChatID, s[0].ID),
			MostRecentMessageAt: conv.MostRecentMessageAt,
			FirstLetters:        firstLetters,
		})
	}
	c.explainedMessagesCache.Set(senderID, cacheItem, ttlcache.DefaultTTL)
	return nil
}
