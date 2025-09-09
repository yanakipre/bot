package internal

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yanakipre/bot/app/telegramsearch/cmd/telegramsearch/internal/embeddings"
	"github.com/yanakipre/bot/app/telegramsearch/cmd/telegramsearch/internal/rootcmd"
	scheduler "github.com/yanakipre/bot/app/telegramsearch/cmd/telegramsearch/internal/sheduler"
	"github.com/yanakipre/bot/app/telegramsearch/cmd/telegramsearch/internal/telegram"
	"github.com/yanakipre/bot/app/telegramsearch/cmd/telegramsearch/internal/userbot"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/staticconfig"
	"github.com/yanakipre/bot/internal/logger"
)

var rootCmd *cobra.Command

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		logger.SetNewGlobalLoggerQuietly(logger.DefaultConfig())
		logger.Error(rootCmd.Context(), fmt.Errorf("error executing root command: %w", err))
	}
	cobra.CheckErr(err)
}

func init() {
	rootCmd = rootcmd.NewRootCmd(func(cmd *cobra.Command, cfg *staticconfig.Config) {
		cmd.AddCommand(telegram.Command(cfg))
		cmd.AddCommand(embeddings.Command(cfg))
		cmd.AddCommand(versionCmd)
		cmd.AddCommand(configgenCmd)
		cmd.AddCommand(userbot.Command(cfg))
		cmd.AddCommand(scheduler.NewSchedulerCmd(cfg))
	})
}
