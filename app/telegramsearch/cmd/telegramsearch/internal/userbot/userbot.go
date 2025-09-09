package userbot

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	ctl2 "github.com/yanakipre/bot/app/telegramsearch/cmd/telegramsearch/internal/ctl"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/controllers/controllerv1"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/staticconfig"
	"github.com/yanakipre/bot/internal/clitooling"
)

var (
	cfg            *staticconfig.Config
	ctl            *controllerv1.Ctl
	CmdsToRegister = []*cobra.Command{
		sessiongen,
		dumpChatHistoryDBCmd,
	}
)

func Init(ctx context.Context, staticConfig *staticconfig.Config) error {
	cfg = staticConfig
	controller, err := ctl2.Init(ctx, staticConfig)
	if err != nil {
		return fmt.Errorf("error in controller init: %w", err)
	}
	ctl = controller
	return nil
}

func Command(cfg *staticconfig.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "userbot",
		Short: "Communicate with userbot.",
		// PersistentPreRun will be executed for any subcommand.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// manually call parent cmd
			if err := clitooling.RunParentPersistentPreRun(cmd, args); err != nil {
				return err
			}
			// We initialize packages with subcommands packages AFTER we configured application and
			// logging.
			// Otherwise, outputting errors and configuring clients is not easy to implement nicely.
			return Init(context.TODO(), cfg)
		},
	}
	cmd.AddCommand(CmdsToRegister...)
	return cmd
}
