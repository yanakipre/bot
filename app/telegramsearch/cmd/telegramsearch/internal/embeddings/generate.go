package embeddings

import (
	"fmt"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/controllers/controllerv1/controllerv1models"

	"github.com/spf13/cobra"
	"github.com/yanakipre/bot/internal/yamlfromstruct"
)

var generate = &cobra.Command{
	Use:   "generate",
	Short: "generate embeddings from previously loaded texts",
	Example: `
Query:

	telegramsearch embeddings generate
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		result, err := ctl.GenerateEmbeddings(ctx, controllerv1models.ReqGenerateEmbeddings{})
		if err != nil {
			return err
		}

		_, err = fmt.Fprint(cmd.OutOrStdout(), yamlfromstruct.Generate(ctx, result))
		return err
	},
}
