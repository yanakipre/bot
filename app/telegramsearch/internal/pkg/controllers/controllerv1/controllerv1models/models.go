package controllerv1models

import (
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/storage/storagemodels"
)

type ReqTryEmbedding struct {
	Input string
}

type EmbeddingResponse struct {
	Text string `yaml:"text"`
}

type RespTryEmbedding struct {
	Result []EmbeddingResponse `yaml:"result"`
}

type ReqTryCompletion struct {
	SenderID int
	Query    string
}

type RespTryCompletion struct {
	Response          string `yaml:"response"`
	UsedConversations []storagemodels.RespSimilaritySearch
}

type ReqGenerateEmbeddings struct {
}

type RespGenerateEmbeddings struct {
}

type ReqDumpChatHistory struct {
	ChatID      string
	ChatHistory []byte
}

type RespDumpChatHistory struct {
}

type ReqDumpChatHistoryFromDB struct {
	ChatID string
}

type RespDumpChatHistoryFromDB struct {
	ThreadsCreated int
}
