package userbot

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/transport/usertransport"
)

var sessiongen = &cobra.Command{
	Use:   "read",
	Short: "read tg chats",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
		defer cancel()

		// Use usertransport with configuration from staticconfig
		if err := usertransport.ReadChatMessages(ctx, cfg.UserTransport, cfg.PostgresRW); err != nil {
			return fmt.Errorf("failed to initialize generate new session: %w", err)
		}

		return nil
	},
}
