package internal

import (
	"github.com/spf13/cobra"
	"github.com/yanakipe/bot/app/telegramsearch/cmd/telegramsearch/internal/embeddings"
	"github.com/yanakipe/bot/app/telegramsearch/cmd/telegramsearch/internal/rootcmd"
	"github.com/yanakipe/bot/app/telegramsearch/cmd/telegramsearch/internal/telegram"
	"github.com/yanakipe/bot/app/telegramsearch/internal/pkg/staticconfig"
	"github.com/yanakipe/bot/internal/logger"
	"go.uber.org/zap"
)

var rootCmd *cobra.Command

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		logger.SetNewGlobalLoggerQuietly(logger.DefaultConfig())
		logger.Error(rootCmd.Context(), "finished with error", zap.Error(err))
	}
	cobra.CheckErr(err)
}

func init() {
	rootCmd = rootcmd.NewRootCmd(func(cmd *cobra.Command, cfg *staticconfig.Config) {
		cmd.AddCommand(telegram.Command(cfg))
		cmd.AddCommand(embeddings.Command(cfg))
		cmd.AddCommand(versionCmd)
		cmd.AddCommand(configgenCmd)
	})
}
