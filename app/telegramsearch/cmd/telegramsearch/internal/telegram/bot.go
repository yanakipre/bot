package telegram

import (
	"fmt"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/transport/bottransport"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

var bot = &cobra.Command{
	Use:   "bot",
	Short: "run bot",
	Example: `
Start bot:

	telegramsearch telegram bot
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
		defer cancel()

		b, err := bottransport.New(ctx, ctl, cfg.TelegramTransport)
		if err != nil {
			return fmt.Errorf("new bot: %w", err)
		}
		//
		//_ = telegram.NewClient(cfg.TelegramV2.AppID.Unmask(), cfg.TelegramV2.AppHash.Unmask(),
		//	telegram.Options{
		//		Logger: logger.FromContext(ctx).Named("tdbot"),
		//	})

		go func() {
			b.Start()
		}()

		<-ctx.Done()

		b.Stop()
		return nil
	},
}
