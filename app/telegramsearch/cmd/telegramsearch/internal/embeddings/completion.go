package embeddings

import (
	"fmt"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/controllers/controllerv1/controllerv1models"

	"github.com/spf13/cobra"
)

var completion = &cobra.Command{
	Use:   "completion",
	Short: "Chat completion",
	Example: `
Query:

	telegramsearch embeddings completion "my question about cyprus"
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		result, err := ctl.TryCompletion(ctx, controllerv1models.ReqTryCompletion{
			Query: args[0],
		})
		if err != nil {
			return err
		}

		_, err = fmt.Fprint(cmd.OutOrStdout(), result.Response)
		return err
	},
}
