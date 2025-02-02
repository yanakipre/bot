package appv1

import (
	"github.com/gotd/td/telegram"
	"github.com/yanakipe/bot/internal/logger"
)

type Deps struct {
	lg     logger.Logger
	client *telegram.Client
}
