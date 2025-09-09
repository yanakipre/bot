package userbot

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	models "github.com/yanakipre/bot/app/telegramsearch/internal/pkg/controllers/controllerv1/controllerv1models"
	"github.com/yanakipre/bot/internal/logger"
	"go.uber.org/zap"
)

var dumpChatHistoryDBCmd = &cobra.Command{
	Use:   "dump-chat-history-db [chat_id]",
	Short: "Dump chat history from database and create threads",
	Long:  "Reads chat messages from database, finds conversation threads and stores them for further processing",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		lg := logger.FromContext(ctx)

		chatID := args[0]
		if chatID == "" {
			return fmt.Errorf("chat_id is required")
		}

		lg.Info("Starting dump chat history from DB", zap.String("chat_id", chatID))

		req := models.ReqDumpChatHistoryFromDB{
			ChatID: chatID,
		}

		resp, err := ctl.DumpChatHistoryFromDB(ctx, req)
		if err != nil {
			lg.Error("Failed to dump chat history from DB", zap.Error(err))
			return fmt.Errorf("failed to dump chat history from DB: %w", err)
		}

		lg.Info("Successfully dumped chat history from DB",
			zap.String("chat_id", chatID),
			zap.Int("threads_created", resp.ThreadsCreated))

		return nil
	},
}
