package controllerv1

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/yanakipe/bot/app/telegramsearch/internal/pkg/client/storage/storagemodels"
	models "github.com/yanakipe/bot/app/telegramsearch/internal/pkg/controllers/controllerv1/controllerv1models"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/samber/lo"
	"github.com/sourcegraph/conc/pool"
	"github.com/yanakipe/bot/internal/logger"
	"go.uber.org/zap"
)

type TextEntity struct {
	Text string `json:"text"`
}

type ChatMessageType string

const (
	ChatMessageTypeMessage ChatMessageType = "message"
)

// Example:
type serializedChatMessage struct {
	ID int64 `json:"id"`
	// Type is service or message
	Type         ChatMessageType `json:"type"`
	DateUnix     string          `json:"date_unixtime"`
	FromId       string          `json:"from_id"`
	TextEntities []TextEntity    `json:"text_entities"`
	text         string          `json:"-"`
	// Reply to some message with ID
	Reply int64 `json:"reply_to_message_id,omitempty"`
}

func (s *serializedChatMessage) getText() string {
	return strings.Join(
		lo.Map(s.TextEntities, func(item TextEntity, _ int) string { return item.Text }),
		" ",
	)
}

type results struct {
	Messages []serializedChatMessage `json:"messages"`
}

type thread []serializedChatMessage

type foundThreads []thread

func (t thread) String() string {
	return strings.Join(lo.Map(t, func(item serializedChatMessage, _ int) string {
		return item.getText()
	}), "\n---->")
}

func (t thread) ForEmbedding() string {
	return strings.Join(lo.Map(t, func(item serializedChatMessage, _ int) string {
		return item.getText()
	}), "\n\n")
}

type response struct {
	Text string
	Date string
}

type data struct {
	// should contain additional context, like Bangkok
	ChatID                string
	ConversationStartedAt string
	ConversationStarter   string
	WithAnswers           bool
	// name of fields that are present in both Config and UserSettingsFeatures
	Responses []response
}

//go:embed for_showing_to_the_user.tmpl
var rawTmpl []byte
var parsedTmpl *template.Template

func init() {
	tmpl, err := template.New("tpl").Parse(string(rawTmpl))
	if err != nil {
		panic(err)
	}
	parsedTmpl = tmpl
}

func processTemplate(data data) (string, error) {
	var processed bytes.Buffer
	err := parsedTmpl.ExecuteTemplate(&processed, "tpl", data)
	if err != nil {
		return "", fmt.Errorf("unable to parse data into template: %w", err)
	}
	return processed.String(), nil
}

func (t thread) ForShowingToTheUser(locality string) (string, error) {
	i, err := strconv.ParseInt(t[0].DateUnix, 10, 64)
	if err != nil {
		panic(err)
	}
	tm := time.Unix(i, 0)
	return processTemplate(data{
		ChatID:                locality,
		ConversationStartedAt: tm.Format("02 Jan 06 15:04"),
		ConversationStarter:   t[0].getText(),
		WithAnswers:           len(t) > 1,
		Responses: lo.Map(t[1:], func(item serializedChatMessage, _ int) response {
			i, err = strconv.ParseInt(item.DateUnix, 10, 64)
			if err != nil {
				panic(err)
			}
			tm = time.Unix(i, 0)
			return response{
				Text: item.getText(),
				Date: tm.Format("02 Jan 06 15:04"),
			}
		}),
	})
}

func findThreads(lg logger.Logger, input []serializedChatMessage) foundThreads {
	r := make(foundThreads, 0, len(input))
	threadIdx := 0
	// map message ID to thread Idx
	mapMsgToThreadIdx := make(map[int64]int, len(input))
	for _, v := range input {
		if v.Reply == 0 {
			// this is a start of thread
			r = append(r, thread{})
			r[threadIdx] = append(r[threadIdx], v)
			mapMsgToThreadIdx[v.ID] = threadIdx
			threadIdx += 1
		} else {
			place, exists := mapMsgToThreadIdx[v.Reply]
			if !exists {
				lg.Debug("message not found, probably deleted, skipped", zap.Int64("id", v.Reply))
				continue
			}
			r[place] = append(r[place], v)
			mapMsgToThreadIdx[v.ID] = place
		}
	}
	return r
}

// high mem consumption path, let gc work
func (c *Ctl) threadsHighMem(ctx context.Context, req models.ReqDumpChatHistory) (foundThreads, error) {
	lg := logger.FromContext(ctx)
	var result results

	if err := json.Unmarshal(req.ChatHistory, &result); err != nil {
		return nil, err
	}
	// leave only "message" type
	msgs := lo.Filter(result.Messages, func(item serializedChatMessage, index int) bool {
		return item.Type == ChatMessageTypeMessage
	})
	threads := lo.Filter(findThreads(lg, msgs), func(item thread, index int) bool {
		return len(item) > 1 // skip threads of len 1 because no answers means no opinions
	})
	r := make(foundThreads, len(threads))
	for i := range threads {
		r[i] = threads[i]
	}
	return r, nil
}

func (c *Ctl) DumpChatHistory(ctx context.Context, req models.ReqDumpChatHistory) (models.RespDumpChatHistory, error) {
	threads, err := c.threadsHighMem(ctx, req)
	if err != nil {
		return models.RespDumpChatHistory{}, fmt.Errorf("unable to get threads: %w", err)
	}
	p := pool.New().WithMaxGoroutines(100).WithContext(ctx)
	for i := range threads {
		t := threads[i]
		p.Go(func(ctx context.Context) error {
			_, err := c.storageRW.CreateChatThread(ctx, storagemodels.ReqCreateChatThread{
				ChatID: storagemodels.ChatID(req.ChatID),
				Body:   t,
			})
			return err
		})
	}
	if err := p.Wait(); err != nil {
		return models.RespDumpChatHistory{}, err
	}

	return models.RespDumpChatHistory{}, nil
}
