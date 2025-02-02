package rootcmd

import (
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/yanakipe/bot/app/telegramsearch/internal/pkg/staticconfig"
	"github.com/yanakipe/bot/internal/config"
	"github.com/yanakipe/bot/internal/logger"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

const (
	defaultConfigName = "config.yaml"
	binaryName        = "telegramsearch"
)

const (
	rootExampleMsg = `
Show available entities:

	telegramsearch --help

	mkdir ~/.telegramsearch/
	yanakipre configgen > ~/.telegramsearch/config.yaml
	# change access rights to this file:
	chmod 0400 ~/.telegramsearch/config.yaml

By default logs are printed to stdout, you can configure sink to some static place.
`
	rootLongDescription = `This binary is intended to manage
the community Q&A project telegramsearch

The binary is organised in manner close to RESTful APIs:
* The first argument is usually a noun representing entity to be operated on.
* The second argument is usually a verb, representing action that will be applied to the entity.
`
)

var (
	// cfgName is Name of config to use
	cfgName string
	cfg     = lo.ToPtr(staticconfig.DefaultConfig())
)

// NewRootCmd represents the base command when called without any subcommands
func NewRootCmd(visit func(*cobra.Command, *staticconfig.Config)) *cobra.Command {
	cmd := &cobra.Command{
		Use:           binaryName,
		Short:         "telegramsearch management tool for administrators.",
		Long:          rootLongDescription,
		Example:       rootExampleMsg,
		SilenceUsage:  true, // do not output usage on error
		SilenceErrors: true, // do not print error twice
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// children of this command will always execute this code by convention
			// so, we initialize all clients here.
			ctx := cmd.Context()

			// do not load config when generating config
			if cmd.Name() == "configgen" {
				return nil
			}

			if cfgName == "" {
				cfgName = defaultConfigName
			}

			err := config.Load(ctx, binaryName, cfg, cfgName)
			if err != nil {
				panic(err)
			}

			logger.SetNewGlobalLoggerOnce(cfg.Logging)

			marshal, err := yaml.Marshal(cfg)
			if err != nil {
				panic(err)
			}
			logger.Debug(ctx, "running with config", zap.ByteString("config", marshal))

			if cmd.Name() == "version" {
				// we don't need to load a profile for particular commands
				return nil
			}

			return nil
		},
	}
	cmd.PersistentFlags().StringVarP(
		&cfgName,
		"config",
		"c",
		defaultConfigName,
		"Use a specific config located in application directory.",
	)
	visit(cmd, cfg)
	return cmd
}
