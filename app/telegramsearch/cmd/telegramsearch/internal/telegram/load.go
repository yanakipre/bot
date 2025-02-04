package telegram

import (
	"fmt"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/controllers/controllerv1/controllerv1models"
	"os"

	"github.com/spf13/cobra"
	"github.com/yanakipre/bot/internal/yamlfromstruct"
)

var chatID *string
var filename *string

var load = &cobra.Command{
	Use:   "load",
	Short: "load exported chat from telegram in json format",
	Example: `
Create a chat and load all messages into it:

	telegramsearch telegram load --filename ~/chat/result.json --chat-id kiprchat --create
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		file, err := os.ReadFile(*filename)
		if err != nil {
			return fmt.Errorf("cannot open file: %w", err)
		}

		chatHistory, err := ctl.DumpChatHistory(ctx, controllerv1models.ReqDumpChatHistory{
			ChatHistory: file,
			ChatID:      *chatID,
		})
		if err != nil {
			return err
		}

		_, err = fmt.Fprint(cmd.OutOrStdout(), yamlfromstruct.Generate(ctx, chatHistory))
		return err
	},
}

func init() {
	chatID = load.Flags().String("chat-id", "", "Chat ID")
	filename = load.Flags().String("filename", "", "Load file")
}
