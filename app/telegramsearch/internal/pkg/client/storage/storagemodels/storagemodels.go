package storagemodels

import "time"

type ReqSimilaritySearch struct {
	// at which distance to not include results
	CutThreshold float32
	Embedding    []float32
	Since        time.Time
	UpTo         time.Time
	Limit        int
}

type RespSimilaritySearch struct {
	// This is message generated for the prompt.
	Message string
	// Telegram chat ID.
	TelegramChatID string
	// This is the original first message of the thread.
	ConversationStarter string
	MostRecentMessageAt time.Time
}

type ReqUpsertEmbedding struct {
	Embedding []float32
	Message   string
	ChatID    string
	ThreadID  int64
}

type RespUpsertEmbedding struct {
}

type ChatID string

type ChatThreadToGenerateEmbedding struct {
	ChatID   string
	ThreadID int64
	Body     []byte
}

type ReqFetchChatThreadToGenerateEmbedding struct {
}

type RespFetchChatThreadToGenerateEmbedding struct {
	Threads []ChatThreadToGenerateEmbedding
}

type ReqCreateChatThread struct {
	ChatID          ChatID
	Body            any
	ThreadStarterID int64
}

type RespCreateChatThread struct {
}

type ReqCreateChat struct {
	ChatID ChatID
}

type RespCreateChat struct {
}

type ReqGetChatLastReadID struct {
	ChatID string
}

type RespGetChatLastReadID struct {
	LastReadID int64
}

type ReqUpdateChatLastReadID struct {
	ChatID     string
	LastReadID int64
}

type RespUpdateChatLastReadID struct {
}

type ReqCreateChatMessage struct {
	MessageID    int64
	ChatID       string
	FromID       int64
	TextEntities string
	Message      string
	ReplyTo      *int64
	Date         time.Time
	Type         string
}

type RespCreateChatMessage struct {
	MessageID int64
}

type ReqFetchChatMessages struct {
	ChatID string
}

type ChatMessage struct {
	MessageID    int64
	ChatID       string
	FromID       int64
	TextEntities string
	Message      string
	ReplyTo      *int64
	Date         time.Time
	Type         string
}

type RespFetchChatMessages struct {
	Messages []ChatMessage
}
