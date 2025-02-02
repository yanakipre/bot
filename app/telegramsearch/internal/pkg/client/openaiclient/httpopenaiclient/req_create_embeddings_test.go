package httpopenaiclient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yanakipe/bot/internal/secret"
	"github.com/yanakipe/bot/internal/testtooling"
	"github.com/yanakipe/bot/telegramsearch/internal/client/openaiclient/openaimodels"
)

func TestClient_CreateEmbeddings(t *testing.T) {
	testtooling.SetNewGlobalLoggerQuietly()

	// Obtain an API Key from https://platform.openai.com/api-keys
	// and replace the empty string to rewrite the requests
	accessKey := secret.NewString(
		"",
	)

	ctx := context.Background()
	type input struct {
		Req openaimodels.ReqCreateEmbeddings
	}
	type output struct {
		Res openaimodels.RespCreateEmbeddings
		Err error
	}
	tests := []struct {
		name string
		when func(ctx context.Context, t *testing.T, i *input)
		then func(t *testing.T, output output)
	}{
		{
			name: "it works",
			when: func(ctx context.Context, t *testing.T, i *input) {
				i.Req = openaimodels.ReqCreateEmbeddings{
					Input: []string{
						"Николай, [15 марта 2023 г., 4:01:43 PM]:\nПодскажите эт сможет бытьплатным или бесплатным паркингом? Вижу вдалеке букву Р но нет паркоматов\n",
						"Alex - Limassol, [15 марта 2023 г., 4:14:37 PM]:\nда, это уличный паркинг. Если под буквой P нет уточняющих надписей или рисунков с монетами, значит бесплатно\n",
						"Николай, [15 марта 2023 г., 4:20:50 PM]:\nАга понял, спасибо, хоть знать буду)\n",
						"Sergey D, [15 марта 2023 г., 5:23:36 PM]:\nа за парковку вообще штрафуют на кипре?)\n",
						"Марина, [15 марта 2023 г., 9:37:55 PM]:\nМеня штрафовали на 50 евро. Случайно припарковалась на стоянке такси возле зоопарка, не заметила надпись\n",
						"Наталья Забелина, [16 марта 2023 г., 9:08:39 AM]:\nДобрый день. А как вам пришел штраф? Или вам его выписали на месте? Спасибо за ответ\n",
						"Марина, [16 марта 2023 г., 9:15:42 AM]:\nЛежал на лобовом стекле\n",
						"Svetlana, [16 марта 2023 г., 10:49:08 AM]:\nОчень даже. Ну и за инвалидное место-очень часто. Так что, туда точно не надо. Вроде 350 евро, если ничего не изменилось в большую сторону)\n",
						"Sergey D, [16 марта 2023 г., 10:49:45 AM]:\nНа инвалидном грех парковаться... я больше про неоплаченную парковку.\n",
					},
				}
			},
			then: func(t *testing.T, output output) {
				require.NoError(t, output.Err)
				require.Len(t, output.Res.Embeddings, 9)
				require.Len(t, output.Res.Embeddings[0].Embedding, 1536)
				// check at least something
				require.Equal(t, output.Res.Embeddings[0].Embedding[:2], []float32{0.002596922, -0.012126352})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cancel, c := ClientWithRecorder(t, tt.name, DefaultConfig(), accessKey)
			defer cancel() // Make sure recorder is stopped once done with it

			i := &input{}
			tt.when(ctx, t, i)

			got, err := c.CreateEmbeddings(ctx, i.Req)

			tt.then(t, output{
				Res: got,
				Err: err,
			})
		})
	}
}
