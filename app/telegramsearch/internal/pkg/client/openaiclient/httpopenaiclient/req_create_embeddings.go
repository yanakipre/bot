package httpopenaiclient

import (
	"context"
	"github.com/yanakipe/bot/app/telegramsearch/internal/pkg/client/openaiclient/openaimodels"

	"github.com/sashabaranov/go-openai"
)

func (c *Client) CreateEmbeddings(ctx context.Context, req openaimodels.ReqCreateEmbeddings) (openaimodels.RespCreateEmbeddings, error) {
	queryReq := openai.EmbeddingRequest{
		Input: req.Input,
		Model: c.cfg.EmbeddingConfig.Model,
	}
	// Create an embedding for the user query
	got, err := c.c.CreateEmbeddings(ctx, queryReq)
	if err != nil {
		return openaimodels.RespCreateEmbeddings{}, handleError(err)
	}
	return openaimodels.RespCreateEmbeddings{
		Embeddings: got.Data,
	}, nil
}
