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
VALUES ($1)
ON CONFLICT (chat_id) DO NOTHING;
`,
	nil,
)

var queryGetChatLastReadID = sqltooling.NewStmt(
	"GetChatLastReadID",
	`
SELECT last_read_id
FROM chats
WHERE chat_id = $1;
`,
	nil,
)

var queryUpdateChatLastReadID = sqltooling.NewStmt(
	"UpdateChatLastReadID",
	`
UPDATE chats
SET last_read_id = $2
WHERE chat_id = $1;
`,
	nil,
)

func (s *Storage) CreateChat(ctx context.Context, req models.ReqCreateChat) (models.RespCreateChat, error) {
	_, err := s.db.ExecContext(ctx, queryCreateChat.Query, req.ChatID)
	if err != nil {
		return models.RespCreateChat{}, err
	}
	return models.RespCreateChat{}, nil
}

func (s *Storage) GetChatLastReadID(ctx context.Context, req models.ReqGetChatLastReadID) (models.RespGetChatLastReadID, error) {
	var lastReadID int64
	err := s.db.GetContext(ctx, &lastReadID, queryGetChatLastReadID.Query, req.ChatID)
	if err != nil {
		return models.RespGetChatLastReadID{}, err
	}
	return models.RespGetChatLastReadID{LastReadID: lastReadID}, nil
}

func (s *Storage) UpdateChatLastReadID(ctx context.Context, req models.ReqUpdateChatLastReadID) (models.RespUpdateChatLastReadID, error) {
	_, err := s.db.ExecContext(ctx, queryUpdateChatLastReadID.Query, req.ChatID, req.LastReadID)
	if err != nil {
		return models.RespUpdateChatLastReadID{}, err
	}
	return models.RespUpdateChatLastReadID{}, nil
}
