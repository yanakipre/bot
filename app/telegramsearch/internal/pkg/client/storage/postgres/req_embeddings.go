package postgres

import (
	"context"
	"github.com/yanakipe/bot/app/telegramsearch/internal/pkg/client/storage/postgres/internal/dbmodels"
	models "github.com/yanakipe/bot/app/telegramsearch/internal/pkg/client/storage/storagemodels"

	"github.com/pgvector/pgvector-go"
	"github.com/samber/lo"
	"github.com/yanakipe/bot/internal/sqltooling"
)

var querySimilaritySearch = sqltooling.NewStmt(
	"SimilaritySearch",
	`
SELECT * FROM
(
	SELECT e.message, e.embedding, t.most_recent_message_at, t.body, c.telegram_chat_id
	FROM embeddings e
		JOIN chatthreads t ON e.thread_id = t.thread_id
		JOIN chats c ON t.chat_id = c.chat_id
	WHERE
		most_recent_message_at >= :since
		AND most_recent_message_at < :upto
	ORDER BY embedding <-> :emb
	LIMIT :limit
) t
`,
	dbmodels.PGSimilarity{},
)

func (s *Storage) FetchSimilaritySearch(ctx context.Context, req models.ReqSimilaritySearch) ([]models.RespSimilaritySearch, error) {
	query, err := s.db.PrepareNamedContext(ctx,
		querySimilaritySearch.Query,
		querySimilaritySearch.Name)
	if err != nil {
		return nil, err
	}

	rows := []dbmodels.PGSimilarity{}
	if err = query.SelectContext(ctx, &rows, map[string]any{
		"threshold": req.CutThreshold,
		"since":     req.Since,
		"upto":      req.UpTo,
		"limit":     req.Limit,
		"emb":       pgvector.NewVector(req.Embedding[:2000]),
	}); err != nil {
		return nil, err
	}
	return lo.Map(rows, func(item dbmodels.PGSimilarity, _ int) models.RespSimilaritySearch {
		return models.RespSimilaritySearch{
			TelegramChatID:      item.TelegramChatID,
			ConversationStarter: item.ConversationStarter,
			Message:             item.Message,
			MostRecentMessageAt: item.MostRecentMessageAt,
		}
	}), nil
}

var queryUpsertEmbedding = sqltooling.NewStmt(
	"UpsertEmbedding",
	`
INSERT INTO embeddings
	(thread_id, chat_id, message, embedding)
VALUES (:thread_id, :chat_id, :message, :embedding)
ON CONFLICT (thread_id) DO UPDATE
	SET
		embedding = EXCLUDED.embedding,
		message = EXCLUDED.message;
`,
	nil,
)

func (s *Storage) UpsertEmbedding(ctx context.Context, req models.ReqUpsertEmbedding) (models.RespUpsertEmbedding, error) {
	query, err := s.db.PrepareNamedContext(ctx,
		queryUpsertEmbedding.Query,
		queryUpsertEmbedding.Name)
	if err != nil {
		return models.RespUpsertEmbedding{}, err
	}

	if _, err = query.ExecContext(ctx, map[string]any{
		"thread_id": req.ThreadID,
		"chat_id":   req.ChatID,
		"message":   req.Message,
		"embedding": pgvector.NewVector(req.Embedding[:2000]),
	}); err != nil {
		return models.RespUpsertEmbedding{}, err
	}
	return models.RespUpsertEmbedding{}, nil
}
