package dbmodels

import "time"

type PGSimilarity struct {
	TelegramChatID      string
	ConversationStarter string `db:"body"`
	Message             string
	MostRecentMessageAt time.Time
}
