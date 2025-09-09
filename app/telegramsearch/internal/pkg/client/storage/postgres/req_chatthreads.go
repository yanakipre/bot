package postgres

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/storage/postgres/internal/dbmodels"
	models "github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/storage/storagemodels"

	"github.com/samber/lo"
	"github.com/yanakipre/bot/internal/sqltooling"
)

var queryFetchChatThreadToGenerateEmbedding = sqltooling.NewStmt(
	"FetchChatThreadToGenerateEmbedding",
	`
SELECT thread_id, chat_id, body FROM chatthreads WHERE most_recent_message_at > fresh_embeddings_at LIMIT 2000;
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
        $1, CAST($2 as JSONB)
);
`,
	nil,
)

func (s *Storage) CreateChatThread(ctx context.Context, req models.ReqCreateChatThread) (models.RespCreateChatThread, error) {
	marshal, err := json.Marshal(req.Body)
	if err != nil {
		return models.RespCreateChatThread{}, err
	}

	_, err = s.db.ExecContext(ctx, queryCreateChatThread.Query, req.ChatID, marshal)
	if err != nil {
		return models.RespCreateChatThread{}, err
	}
	return models.RespCreateChatThread{}, nil
}

var queryCreateChatThreadNew = sqltooling.NewStmt(
	"CreateChatThreadNew",
	`
INSERT INTO chatthreads
	(chat_id, body, thread_starter_id, most_recent_message_at)
VALUES (
        $1, CAST($2 as JSONB), $3, NOW()
)
ON CONFLICT (chat_id, thread_starter_id) DO UPDATE SET
    body = CAST($2 as JSONB),
    most_recent_message_at = CASE 
        WHEN chatthreads.body IS DISTINCT FROM CAST($2 as JSONB) THEN NOW() 
        ELSE chatthreads.most_recent_message_at 
    END;
`,
	nil,
)

func (s *Storage) CreateChatThreadNew(ctx context.Context, req models.ReqCreateChatThread) (models.RespCreateChatThread, error) {
	marshal, err := json.Marshal(req.Body)
	if err != nil {
		return models.RespCreateChatThread{}, err
	}

	_, err = s.db.ExecContext(ctx, queryCreateChatThreadNew.Query, req.ChatID, marshal, strconv.FormatInt(req.ThreadStarterID, 10))
	if err != nil {
		return models.RespCreateChatThread{}, err
	}
	return models.RespCreateChatThread{}, nil
}
