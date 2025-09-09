package userbot

import (
	"context"
	"fmt"
	"time"

	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/openaiclient/httpopenaiclient"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/storage/postgres"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/controllers/controllerv1"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/controllers/controllerv1/controllerv1models"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/staticconfig"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/transport/usertransport"
	"github.com/yanakipre/bot/internal/encodingtooling"
	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/scheduletooling"
	"github.com/yanakipre/bot/internal/scheduletooling/worker"
)

func NewChatReaderJob(cfg *staticconfig.Config) scheduletooling.Job {
	return scheduletooling.NewConcurrentInProcessJobWithCloser(
		func(ctx context.Context) error {
			logger.Info(ctx, "Starting scheduled chat reader job")

			err := usertransport.ReadChatMessages(ctx, cfg.UserTransport, cfg.PostgresRW)
			if err != nil {
				logger.Error(ctx, fmt.Errorf("chat reader job failed: %w", err))
				return err
			}

			logger.Info(ctx, "Chat reader job completed successfully")
			return nil
		},
		nil,
		scheduletooling.Config{
			UniqueName: "telegram_chat_reader",
			Enabled:    true,
			Interval:   encodingtooling.Duration{Duration: 15 * time.Second},
			Timeout:    encodingtooling.Duration{Duration: 5 * time.Minute},
		},
		scheduletooling.ConstantConfig(scheduletooling.Config{
			UniqueName: "telegram_chat_reader",
			Enabled:    true,
			Interval:   encodingtooling.Duration{Duration: 15 * time.Second},
			Timeout:    encodingtooling.Duration{Duration: 5 * time.Minute},
		}),
		worker.NewWellKnownMetricsCollector(),
		1,
	)
}

func NewChatHistoryDumpJob(cfg *staticconfig.Config) scheduletooling.Job {
	return scheduletooling.NewConcurrentInProcessJobWithCloser(
		func(ctx context.Context) error {
			logger.Info(ctx, "Starting scheduled chat history dump job")

			storageRW := postgres.New(cfg.PostgresRW)
			err := storageRW.Ready(ctx)
			if err != nil {
				return fmt.Errorf("error creating storage: %w", err)
			}
			defer storageRW.Close()
			openai := httpopenaiclient.NewClient(cfg.OpenAI)

			ctl, err := controllerv1.New(cfg.Ctlv1, openai, storageRW)
			if err != nil {
				return fmt.Errorf("error creating controller: %w", err)
			}

			if len(cfg.UserTransport.ChatIDs) == 0 {
				return fmt.Errorf("no chat IDs configured")
			}

			for _, chatIDSecret := range cfg.UserTransport.ChatIDs {
				chatID := chatIDSecret.Unmask()
				req := controllerv1models.ReqDumpChatHistoryFromDB{
					ChatID: fmt.Sprintf("%d", chatID),
				}

				if _, err := ctl.DumpChatHistoryFromDB(ctx, req); err != nil {
					logger.Error(ctx, fmt.Errorf("chat history dump job failed: %w", err))
					return err
				}
			}

			logger.Info(ctx, "Chat history dump job completed successfully")
			return nil
		},
		nil,
		scheduletooling.Config{
			UniqueName: "telegram_chat_history_dump",
			Enabled:    true,
			Interval:   encodingtooling.Duration{Duration: 15 * time.Second},
			Timeout:    encodingtooling.Duration{Duration: 10 * time.Minute},
		},
		scheduletooling.ConstantConfig(scheduletooling.Config{
			UniqueName: "telegram_chat_history_dump",
			Enabled:    true,
			Interval:   encodingtooling.Duration{Duration: 15 * time.Second},
			Timeout:    encodingtooling.Duration{Duration: 10 * time.Minute},
		}),
		worker.NewWellKnownMetricsCollector(),
		1,
	)
}

func NewEmbeddingsGenerationJob(cfg *staticconfig.Config) scheduletooling.Job {
	return scheduletooling.NewConcurrentInProcessJobWithCloser(
		func(ctx context.Context) error {
			logger.Info(ctx, "Starting scheduled embeddings generation job")

			storageRW := postgres.New(cfg.PostgresRW)
			if err := storageRW.Ready(ctx); err != nil {
				return fmt.Errorf("error creating storage: %w", err)
			}
			defer storageRW.Close()
			openai := httpopenaiclient.NewClient(cfg.OpenAI)

			ctl, err := controllerv1.New(cfg.Ctlv1, openai, storageRW)
			if err != nil {
				return fmt.Errorf("error creating controller: %w", err)
			}

			_, err = ctl.GenerateEmbeddings(ctx, controllerv1models.ReqGenerateEmbeddings{})
			if err != nil {
				logger.Error(ctx, fmt.Errorf("embeddings generation job failed: %w", err))
				return err
			}

			logger.Info(ctx, "Embeddings generation job completed successfully")
			return nil
		},
		nil,
		scheduletooling.Config{
			UniqueName: "telegram_embeddings_generate",
			Enabled:    true,
			Interval:   encodingtooling.Duration{Duration: 3000 * time.Minute},
			Timeout:    encodingtooling.Duration{Duration: 20 * time.Minute},
		},
		scheduletooling.ConstantConfig(scheduletooling.Config{
			UniqueName: "telegram_embeddings_generate",
			Enabled:    true,
			Interval:   encodingtooling.Duration{Duration: 3000 * time.Minute},
			Timeout:    encodingtooling.Duration{Duration: 20 * time.Minute},
		}),
		worker.NewWellKnownMetricsCollector(),
		1,
	)
}
