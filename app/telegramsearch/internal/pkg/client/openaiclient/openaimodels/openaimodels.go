package openaimodels

import "github.com/sashabaranov/go-openai"

type ReqCreateChatCompletion struct {
	Input         string
	Conversations []string
}

type RespCreateChatCompletion struct {
	Response string
}

type ReqCreateEmbeddings struct {
	Input []string
}

type RespCreateEmbeddings struct {
	Embeddings []openai.Embedding
}
