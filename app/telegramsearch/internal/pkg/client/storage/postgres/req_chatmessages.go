package postgres

import (
	"context"

	models "github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/storage/storagemodels"

	"github.com/yanakipre/bot/internal/sqltooling"
)

var queryCreateChatMessage = sqltooling.NewStmt(
	"CreateChatMessage",
	`
WITH ins_chat AS (
    INSERT INTO chats (chat_id)
    VALUES (:chat_id)
    ON CONFLICT (chat_id) DO NOTHING
)
INSERT INTO chatmessages (message_id, chat_id, from_id, text_entities, message, reply_to, date, type)
VALUES (
        :message_id,
        :chat_id,
        :from_id, 
        :text_entities,
        :message,
        :reply_to,
        :date,
        :type
)
ON CONFLICT (message_id, chat_id) DO NOTHING
`,
	nil,
)

func (s *Storage) CreateChatMessage(ctx context.Context, req models.ReqCreateChatMessage) (models.RespCreateChatMessage, error) {
	_, err := s.db.Exec(ctx, queryCreateChatMessage.Name, queryCreateChatMessage.Query, map[string]any{
		"message_id":    req.MessageID,
		"chat_id":       req.ChatID,
		"from_id":       req.FromID,
		"text_entities": req.TextEntities,
		"message":       req.Message,
		"reply_to":      req.ReplyTo,
		"date":          req.Date,
		"type":          req.Type,
	})
	if err != nil {
		return models.RespCreateChatMessage{}, err
	}
	return models.RespCreateChatMessage{MessageID: req.MessageID}, nil
}

var queryFetchChatMessages = sqltooling.NewStmt(
	"FetchChatMessages",
	`
SELECT message_id, chat_id, from_id, text_entities, message, reply_to, date, type
FROM chatmessages 
WHERE chat_id = $1
ORDER BY date ASC, message_id ASC;
`,
	nil,
)

func (s *Storage) FetchChatMessages(ctx context.Context, req models.ReqFetchChatMessages) (models.RespFetchChatMessages, error) {
	var messages []models.ChatMessage
	err := s.db.SelectContext(ctx, &messages, queryFetchChatMessages.Query, req.ChatID)
	if err != nil {
		return models.RespFetchChatMessages{}, err
	}
	return models.RespFetchChatMessages{Messages: messages}, nil
}
