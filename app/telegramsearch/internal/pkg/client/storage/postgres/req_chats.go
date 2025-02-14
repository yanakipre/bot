package postgres

import (
	"context"
	models "github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/storage/storagemodels"

	"github.com/yanakipre/bot/internal/sqltooling"
)

var queryCreateChat = sqltooling.NewStmt(
	"CreateChat",
	`
INSERT INTO chats (chat_id)
VALUES (
        :chat_id
);
`,
	nil,
)

func (s *Storage) CreateChat(ctx context.Context, req models.ReqCreateChat) (models.RespCreateChat, error) {
	_, err := s.db.ExecContext(ctx, queryCreateChat.Query, map[string]any{
		"chat_id": req.ChatID,
	})
	if err != nil {
		return models.RespCreateChat{}, err
	}
	return models.RespCreateChat{}, nil
}
