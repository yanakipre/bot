package controllerv1

import (
	"context"
	_ "embed"
	"testing"

	models "github.com/yanakipre/bot/app/telegramsearch/internal/pkg/controllers/controllerv1/controllerv1models"

	"github.com/stretchr/testify/require"
	"github.com/yanakipre/bot/internal/logger"
)

//go:embed fixtures/cylimassol.json
var history []byte

func TestCtl_DumpChatHistory(t *testing.T) {
	cfg := logger.DefaultConfig()
	cfg.Format = logger.FormatConsole
	cfg.LogLevel = "INFO"
	logger.SetNewGlobalLoggerQuietly(cfg)
	c := Ctl{}
	ctx := context.Background()
	_, err := c.DumpChatHistory(ctx, models.ReqDumpChatHistory{
		ChatHistory: history,
	})
	require.NoError(t, err)

}
