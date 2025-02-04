package appv1

import (
	"github.com/gotd/td/telegram"
	"github.com/yanakipre/bot/internal/logger"
)

type Deps struct {
	lg     logger.Logger
	client *telegram.Client
}
