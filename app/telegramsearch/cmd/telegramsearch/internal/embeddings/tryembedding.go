package embeddings

import (
	"fmt"
	"github.com/yanakipe/bot/app/telegramsearch/internal/pkg/controllers/controllerv1/controllerv1models"

	"github.com/spf13/cobra"
)

var try = &cobra.Command{
	Use:   "try",
	Short: "try matching text to existing embeddings",
	Example: `
Query:

	telegramsearch embeddings try "Are there any chats about limassol?"
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		result, err := ctl.TryEmbedding(ctx, controllerv1models.ReqTryEmbedding{Input: args[0]})
		if err != nil {
			return err
		}

		for _, r := range result.Result {
			if _, err = fmt.Fprint(cmd.OutOrStdout(), r.Text); err != nil {
				return err
			}
		}
		return nil
	},
}
