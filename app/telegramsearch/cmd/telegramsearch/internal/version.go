package internal

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yanakipre/bot/internal/buildtooling"
	"github.com/yanakipre/bot/internal/yamlfromstruct"
)

// configgenCmd represents the init command
var versionCmd = &cobra.Command{
	Use: "version",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := fmt.Fprint(
			cmd.OutOrStdout(),
			yamlfromstruct.Generate(cmd.Context(), buildtooling.Build),
		)
		return err
	},
}
