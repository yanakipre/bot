package controllerv1

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/storage/storagemodels"
	models "github.com/yanakipre/bot/app/telegramsearch/internal/pkg/controllers/controllerv1/controllerv1models"

	"github.com/samber/lo"
	"github.com/sourcegraph/conc/pool"
	"github.com/yanakipre/bot/internal/logger"
	"go.uber.org/zap"
)

type dbChatMessage struct {
	MessageID    int64
	ChatID       string
	FromID       int64
	TextEntities string
	Message      string
	ReplyTo      *int64
	Date         time.Time
	Type         string
}

func (d *dbChatMessage) getText() string {
	if d.Message != "" {
		return d.Message
	}
	var entities []TextEntity
	if err := json.Unmarshal([]byte(d.TextEntities), &entities); err != nil {
		return d.TextEntities
	}
	return strings.Join(
		lo.Map(entities, func(item TextEntity, _ int) string { return item.Text }),
		" ",
	)
}

type dbThread []dbChatMessage
type dbFoundThreads []dbThread

func (t dbThread) String() string {
	return strings.Join(lo.Map(t, func(item dbChatMessage, _ int) string {
		return item.getText()
	}), "\n---->")
}

func (t dbThread) ForEmbedding() string {
	return strings.Join(lo.Map(t, func(item dbChatMessage, _ int) string {
		return item.getText()
	}), "\n\n")
}

func (t dbThread) ForShowingToTheUser(locality string) (string, error) {
	return processTemplate(data{
		ChatID:                locality,
		ConversationStartedAt: t[0].Date.Format("02 Jan 06 15:04"),
		ConversationStarter:   t[0].getText(),
		WithAnswers:           len(t) > 1,
		Responses: lo.Map(t[1:], func(item dbChatMessage, _ int) response {
			return response{
				Text: item.getText(),
				Date: item.Date.Format("02 Jan 06 15:04"),
			}
		}),
	})
}

func findDBThreads(lg logger.Logger, input []dbChatMessage) dbFoundThreads {
	r := make(dbFoundThreads, 0, len(input))
	threadIdx := 0
	mapMsgToThreadIdx := make(map[int64]int, len(input))
	for _, v := range input {
		if v.ReplyTo == nil || *v.ReplyTo == 0 {
			r = append(r, dbThread{})
			r[threadIdx] = append(r[threadIdx], v)
			mapMsgToThreadIdx[v.MessageID] = threadIdx
			threadIdx += 1
		} else {
			place, exists := mapMsgToThreadIdx[*v.ReplyTo]
			if !exists {
				lg.Debug("message not found, probably deleted, skipped", zap.Int64("id", *v.ReplyTo))
				continue
			}
			r[place] = append(r[place], v)
			mapMsgToThreadIdx[v.MessageID] = place
		}
	}
	return r
}

func (c *Ctl) threadsFromDB(ctx context.Context, req models.ReqDumpChatHistoryFromDB) (dbFoundThreads, error) {
	lg := logger.FromContext(ctx)

	resp, err := c.storageRW.FetchChatMessages(ctx, storagemodels.ReqFetchChatMessages{
		ChatID: req.ChatID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to fetch messages from DB: %w", err)
	}

	lg.Info("Fetched messages from DB", zap.Int("total_messages", len(resp.Messages)))

	if len(resp.Messages) > 0 {
		types := lo.Map(resp.Messages[:lo.Min([]int{5, len(resp.Messages)})], func(item storagemodels.ChatMessage, _ int) string {
			return item.Type
		})
		lg.Info("Sample message types", zap.Strings("types", types))
	}

	msgs := lo.Filter(lo.Map(resp.Messages, func(item storagemodels.ChatMessage, _ int) dbChatMessage {
		return dbChatMessage{
			MessageID:    item.MessageID,
			ChatID:       item.ChatID,
			FromID:       item.FromID,
			TextEntities: item.TextEntities,
			Message:      item.Message,
			ReplyTo:      item.ReplyTo,
			Date:         item.Date,
			Type:         item.Type,
		}
	}), func(item dbChatMessage, index int) bool {
		return item.Type == "message" || item.Type == "text" || item.Type == "media"
	})

	lg.Info("Filtered messages by type", zap.Int("message_type_count", len(msgs)))

	allThreads := findDBThreads(lg, msgs)
	lg.Info("Found all threads", zap.Int("all_threads_count", len(allThreads)))

	threads := lo.Filter(allThreads, func(item dbThread, index int) bool {
		return len(item) > 1
	})

	lg.Info("Filtered threads with length > 1", zap.Int("final_threads_count", len(threads)))

	r := make(dbFoundThreads, len(threads))
	for i := range threads {
		r[i] = threads[i]
	}
	return r, nil
}

func (c *Ctl) DumpChatHistoryFromDB(ctx context.Context, req models.ReqDumpChatHistoryFromDB) (models.RespDumpChatHistoryFromDB, error) {
	threads, err := c.threadsFromDB(ctx, req)
	if err != nil {
		return models.RespDumpChatHistoryFromDB{}, fmt.Errorf("unable to get threads from DB: %w", err)
	}

	p := pool.New().WithMaxGoroutines(100).WithContext(ctx)
	for i := range threads {
		t := threads[i]
		p.Go(func(ctx context.Context) error {
			threadData := lo.Map(t, func(item dbChatMessage, _ int) serializedChatMessage {
				var replyID int64
				if item.ReplyTo != nil {
					replyID = *item.ReplyTo
				}

				msgType := item.Type
				if msgType == "text" || msgType == "media" {
					msgType = "message"
				}

				var entities []TextEntity
				if err := json.Unmarshal([]byte(item.TextEntities), &entities); err != nil {
					entities = []TextEntity{{Text: item.TextEntities}}
				}

				return serializedChatMessage{
					ID:           item.MessageID,
					Type:         ChatMessageType(msgType),
					DateUnix:     strconv.FormatInt(item.Date.Unix(), 10),
					FromId:       strconv.FormatInt(item.FromID, 10),
					TextEntities: entities,
					Reply:        replyID,
				}
			})

			_, err := c.storageRW.CreateChatThreadNew(ctx, storagemodels.ReqCreateChatThread{
				ChatID:          storagemodels.ChatID(req.ChatID),
				Body:            threadData,
				ThreadStarterID: t[0].MessageID,
			})
			return err
		})
	}
	if err := p.Wait(); err != nil {
		return models.RespDumpChatHistoryFromDB{}, err
	}

	return models.RespDumpChatHistoryFromDB{ThreadsCreated: len(threads)}, nil
}
