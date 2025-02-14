package postgres

import (
	"context"
	"encoding/json"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/storage/postgres/internal/dbmodels"
	models "github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/storage/storagemodels"

	"github.com/samber/lo"
	"github.com/yanakipre/bot/internal/sqltooling"
)

var queryFetchChatThreadToGenerateEmbedding = sqltooling.NewStmt(
	"FetchChatThreadToGenerateEmbedding",
	`
SELECT * FROM chatthreads WHERE thread_id NOT IN (SELECT thread_id FROM embeddings) LIMIT 2000;
`,
	dbmodels.ChatThread{},
)

func (s *Storage) FetchChatThreadToGenerateEmbedding(ctx context.Context, req models.ReqFetchChatThreadToGenerateEmbedding) (models.RespFetchChatThreadToGenerateEmbedding, error) {
	rows := []dbmodels.ChatThread{}
	if err := s.db.SelectContext(ctx, &rows, queryFetchChatThreadToGenerateEmbedding.Query, map[string]any{}); err != nil {
		return models.RespFetchChatThreadToGenerateEmbedding{}, err
	}
	return models.RespFetchChatThreadToGenerateEmbedding{
		Threads: lo.Map(rows, func(item dbmodels.ChatThread, _ int) models.ChatThreadToGenerateEmbedding {
			return models.ChatThreadToGenerateEmbedding{
				ChatID:   item.ChatID,
				ThreadID: item.ThreadID,
				Body:     item.Body,
			}
		}),
	}, nil
}

var queryCreateChatThread = sqltooling.NewStmt(
	"CreateChatThread",
	`
INSERT INTO chatthreads
	(chat_id, body)
VALUES (
        :chat_id, CAST(:body as JSONB)
);
`,
	nil,
)

func (s *Storage) CreateChatThread(ctx context.Context, req models.ReqCreateChatThread) (models.RespCreateChatThread, error) {
	marshal, err := json.Marshal(req.Body)
	if err != nil {
		return models.RespCreateChatThread{}, err
	}

	_, err = s.db.ExecContext(ctx, queryCreateChatThread.Query, map[string]any{
		"chat_id": req.ChatID,
		"body":    marshal,
	})
	if err != nil {
		return models.RespCreateChatThread{}, err
	}
	return models.RespCreateChatThread{}, nil
}
