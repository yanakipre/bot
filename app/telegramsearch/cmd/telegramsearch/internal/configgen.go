package internal

import (
	"fmt"

	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/staticconfig"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const configgenCmdName = "configgen"

// configgenCmd represents the init command
var configgenCmd = &cobra.Command{
	Use:   configgenCmdName,
	Short: "Generate example config to STDOUT.",
	Long: `We generate config with default staging profile, NOT with production.
Thus production would not be destroyed unintentionally
by running command without --profile=production (or similar).

After generating the config you should supply it with api key by yourself.
`,
	Example: `
Generate config and put it into default place:

	telegramsearch configgen > ~/.telegramsearch/config.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		marshalled, err := yaml.Marshal(exampleConfigForUsers())
		if err != nil {
			return fmt.Errorf("could not marshal config: %w", err)
		}
		fmt.Print(string(marshalled))
		return nil
	},
}

func exampleConfigForUsers() staticconfig.Config {
	cfg := staticconfig.Config{}
	cfg.DefaultConfig()
	return cfg
}
