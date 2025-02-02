package bottransport

import (
	"context"
	"github.com/yanakipe/bot/app/telegramsearch/internal/pkg/controllers/controllerv1"
	"github.com/yanakipe/bot/app/telegramsearch/internal/pkg/controllers/controllerv1/controllerv1models"
	"time"

	"github.com/tucnak/telebot"
	"github.com/yanakipe/bot/internal/logger"
	"go.uber.org/zap"
)

func New(ctx context.Context, ctl *controllerv1.Ctl, cfg Config) (*telebot.Bot, error) {
	lg := logger.FromContext(ctx)
	pref := telebot.Settings{
		Token:  cfg.Token.Unmask(),
		Poller: &telebot.LongPoller{Timeout: 1 * time.Second},
		Reporter: func(err error) {
			lg.Error("Telegram bot failed", zap.Error(err))
		},
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		return nil, err
	}

	b.Handle("/start", func(m *telebot.Message) {
		lg.Info("user joined")
		_, err := b.Send(m.Sender, cfg.Greeting)
		if err != nil {
			lg.Error("Sender failed", zap.Error(err))
			return
		}
	})
	//b.Handle(telebot.OnChannelPost, func(m *telebot.Message) {
	//	logger.Warn(ctx, "channel post", zap.String("text", m.Text))
	//	if strings.Contains(m.Text, "@yanakipre_bot") && m.ReplyTo != nil {
	//		// completion requested
	//		prompt := m.ReplyTo.Text
	//		ctx, cancel := context.WithTimeout(ctx, 100*time.Second)
	//		defer cancel()
	//		completion, err := ctl.TryCompletion(ctx, controllerv1models.ReqTryCompletion{Query: prompt})
	//		if err != nil {
	//			lg.Error("Completion failed", zap.Error(err))
	//			return
	//		}
	//		_, err = b.Send(m.Sender, completion.Response, &telebot.SendOptions{
	//			ReplyTo: m,
	//			//ReplyMarkup:           telebot.ModeMarkdown,
	//			DisableWebPagePreview: false,
	//			//DisableNotification:   false,
	//			ParseMode: telebot.ModeMarkdown,
	//		})
	//		if err != nil {
	//			lg.Error("Sender failed 111", zap.Error(err))
	//			return
	//		}
	//	}
	//})

	b.Handle(telebot.OnText, func(m *telebot.Message) {
		ctx, cancel := context.WithTimeout(ctx, 100*time.Second)
		defer cancel()
		if m.Chat.Type != telebot.ChatPrivate {
			// public chat
			if len(m.Entities) > 0 && m.Entities[0].Type == telebot.EntityMention {
				prompt := m.ReplyTo.Text
				logger.Info(ctx, "mention", zap.String("text", prompt))
				completion, err := ctl.TryCompletion(ctx, controllerv1models.ReqTryCompletion{
					SenderID: 0, // we don't log messages for public chats.
					Query:    prompt,
				})
				if err != nil {
					lg.Error("Completion failed", zap.Error(err))
					return
				}
				_, err = b.Reply(m.ReplyTo, completion.Response, &telebot.SendOptions{
					ReplyTo: m.ReplyTo,
					//ReplyMarkup:           telebot.ModeMarkdown,
					DisableWebPagePreview: true,
					//DisableNotification:   false,
					ParseMode: telebot.ModeDefault,
				})
				if err != nil {
					lg.Error("Failed chat prompt", zap.Error(err))
					return
				}
			}
			return
		}
		if m.Text == explain {
			message, err := ctl.ExplainMessage(ctx, m.Sender.ID)
			if err != nil {
				lg.Error("ExplainMessage", zap.Error(err))
				return
			}
			_, err = b.Send(m.Sender, message, &telebot.SendOptions{
				ReplyTo:               m,
				DisableWebPagePreview: true,
				DisableNotification:   true,
				ParseMode:             telebot.ModeDefault,
			})
			if err != nil {
				lg.Error("explain help", zap.Error(err))
				return
			}
			return
		}
		if m.Text == help {
			_, err = b.Send(m.Sender, ctl.Help(ctx), &telebot.SendOptions{
				ReplyTo:               m,
				DisableWebPagePreview: false,
				DisableNotification:   true,
				ParseMode:             telebot.ModeMarkdown,
			})
			if err != nil {
				lg.Error("Sender help", zap.Error(err))
				return
			}
			return
		}
		if m.Text == news {
			_, err = b.Send(m.Sender, ctl.News(ctx), &telebot.SendOptions{
				ReplyTo:               m,
				DisableWebPagePreview: true,
				DisableNotification:   true,
				ParseMode:             telebot.ModeDefault,
			})
			if err != nil {
				lg.Error("Sender news", zap.Error(err))
				return
			}
			return
		}
		//_, err := b.Send(m.Sender, telebot.Typing)
		//if err != nil {
		//	return
		//}
		completion, err := ctl.TryCompletion(ctx, controllerv1models.ReqTryCompletion{
			SenderID: m.Sender.ID,
			Query:    m.Text,
		})
		if err != nil {
			lg.Error("Completion failed", zap.Error(err))
			return
		}
		_, err = b.Reply(m, completion.Response, &telebot.SendOptions{
			ReplyTo: m,
			//ReplyMarkup:           telebot.ModeMarkdown,
			DisableWebPagePreview: true,
			//DisableNotification:   false,
			ParseMode: telebot.ModeDefault,
		})
		//_, err = b.Send(m.Sender, completion.Response, &telebot.SendOptions{
		//	ReplyTo: m,
		//	//ReplyMarkup:           telebot.ModeMarkdown,
		//	DisableWebPagePreview: true,
		//	//DisableNotification:   false,
		//	ParseMode: telebot.ModeMarkdown,
		//})
		if err != nil {
			lg.Error("Sender failed", zap.Error(err))
			return
		}
		//for i := range completion.UsedConversations {
		//	_, err = b.Send(m.Sender, "used: "+completion.UsedConversations[i].Message, &telebot.SendOptions{
		//		ReplyTo: m,
		//		//ReplyMarkup:           telebot.ModeMarkdown,
		//		DisableWebPagePreview: true,
		//		//DisableNotification:   false,
		//		ParseMode: telebot.ModeMarkdown,
		//	})
		//}
	})
	return b, err
}

const (
	help    = "/help"
	news    = "/news"
	explain = "/explain"
)
