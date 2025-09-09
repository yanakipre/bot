package usertransport

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
	"unicode/utf16"

	"github.com/go-faster/errors"
	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/td/examples"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"golang.org/x/time/rate"

	"go.uber.org/zap"

	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/storage/postgres"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/storage/storagemodels"
	"github.com/yanakipre/bot/internal/logger"
)

func extractEntityText(text string, offset, length int) string {
	runes := []rune(text)
	utf16Codes := utf16.Encode(runes)

	if offset < 0 || offset >= len(utf16Codes) {
		return ""
	}

	endPos := offset + length
	if endPos > len(utf16Codes) {
		endPos = len(utf16Codes)
	}

	if offset >= endPos || length <= 0 {
		return ""
	}

	entityCodeUnits := utf16Codes[offset:endPos]
	entityRunes := utf16.Decode(entityCodeUnits)
	return string(entityRunes)
}

func baseEntity(entityType, message string, offset, length int) map[string]interface{} {
	return map[string]interface{}{
		"type":   entityType,
		"offset": offset,
		"length": length,
		"text":   extractEntityText(message, offset, length),
	}
}

func reconstructFullTextEntities(fullText string, entities []map[string]interface{}) []map[string]interface{} {
	if len(entities) == 0 {
		return []map[string]interface{}{{"text": fullText}}
	}

	result := []map[string]interface{}{}

	for _, entity := range entities {
		result = append(result, entity)
	}

	return result
}

func writeMessagesToDatabase(ctx context.Context, storage *postgres.Storage, messages []*tg.Message, chatIdentifier string) error {
	logger.Debug(ctx, "Saving messages to database", zap.Int("count", len(messages)), zap.String("chat", chatIdentifier))

	for _, msg := range messages {

		var fromID int64 = 0
		if msg.FromID != nil {
			switch p := msg.FromID.(type) {
			case *tg.PeerUser:
				fromID = p.UserID
			case *tg.PeerChat:
				fromID = p.ChatID
			case *tg.PeerChannel:
				fromID = p.ChannelID
			}
		}
		// API returns raw text (while manual export from tg not), but i transform entities to match manual export format
		// to ensure compatibility with existing embedding generation (untestable in current setup)
		enrichedEntities := make([]map[string]interface{}, 0, len(msg.Entities))
		for _, entity := range msg.Entities {
			entityMap := make(map[string]interface{})

			switch e := entity.(type) {
			case *tg.MessageEntityMention:
				entityMap = baseEntity("mention", msg.Message, e.Offset, e.Length)

			case *tg.MessageEntityMentionName:
				entityMap = baseEntity("mention_name", msg.Message, e.Offset, e.Length)
				entityMap["user_id"] = e.UserID

			case *tg.MessageEntityHashtag:
				entityMap = baseEntity("hashtag", msg.Message, e.Offset, e.Length)

			case *tg.MessageEntityURL:
				entityMap = baseEntity("url", msg.Message, e.Offset, e.Length)

			case *tg.MessageEntityBold:
				entityMap = baseEntity("bold", msg.Message, e.Offset, e.Length)

			case *tg.MessageEntityItalic:
				entityMap = baseEntity("italic", msg.Message, e.Offset, e.Length)

			case *tg.MessageEntityCode:
				entityMap = baseEntity("code", msg.Message, e.Offset, e.Length)

			case *tg.MessageEntityBlockquote:
				entityMap = baseEntity("blockquote", msg.Message, e.Offset, e.Length)
				entityMap["collapsed"] = e.Collapsed

			case *tg.MessageEntityTextURL:
				entityMap = baseEntity("text_url", msg.Message, e.Offset, e.Length)
				entityMap["url"] = e.URL

			case *tg.MessageEntitySpoiler:
				entityMap = baseEntity("spoiler", msg.Message, e.Offset, e.Length)

			case *tg.MessageEntityUnderline:
				entityMap = baseEntity("underline", msg.Message, e.Offset, e.Length)

			case *tg.MessageEntityStrike:
				entityMap = baseEntity("strike", msg.Message, e.Offset, e.Length)

			case *tg.MessageEntityPre:
				entityMap = baseEntity("pre", msg.Message, e.Offset, e.Length)
				entityMap["language"] = e.Language

			default:
				entityMap["type"] = "unknown"
			}

			enrichedEntities = append(enrichedEntities, entityMap)
		}

		fullTextEntities := reconstructFullTextEntities(msg.Message, enrichedEntities)
		textEntitiesJSON, err := json.Marshal(fullTextEntities)

		if err != nil {
			return fmt.Errorf("failed to marshal text entities for message %d: %v", msg.ID, err)
		}

		var replyTo *int64
		if msg.ReplyTo != nil {
			switch r := msg.ReplyTo.(type) {
			case *tg.MessageReplyHeader:
				if r.ReplyToMsgID != 0 {
					replyToVal := int64(r.ReplyToMsgID)
					replyTo = &replyToVal
				}
			}
		}

		msgType := "message"
		if msg.Media != nil {
			msgType = "media"
		}
		if msg.Post {
			msgType = "channel_post"
		}

		req := storagemodels.ReqCreateChatMessage{
			MessageID:    int64(msg.ID),
			ChatID:       chatIdentifier,
			FromID:       fromID,
			TextEntities: string(textEntitiesJSON),
			Message:      msg.Message,
			ReplyTo:      replyTo,
			Date:         time.Unix(int64(msg.Date), 0),
			Type:         msgType,
		}

		resp, err := storage.CreateChatMessage(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to save message %d to database: %v", msg.ID, err)
		}

		logger.Debug(ctx, "Saved message", zap.Int("msg_id", msg.ID), zap.Int64("db_id", resp.MessageID))
	}

	logger.Debug(ctx, "Successfully saved messages to database", zap.Int("count", len(messages)))
	return nil
}

func ReadChatMessages(ctx context.Context, cfg Config, postgresConfig postgres.Config) error {

	phone := cfg.Phone.Unmask()
	appID := cfg.AppID.Unmask()
	appHash := cfg.AppHash.Unmask()
	sessionDir := cfg.SessionStoragePath
	// Setting up session storage.
	// This is needed to reuse session and not login every time.
	if err := os.MkdirAll(sessionDir, 0700); err != nil {
		return err
	}

	// So, we are storing session information in current directory, under subdirectory "session/phone_hash"
	sessionStorage := &telegram.FileSessionStorage{
		Path: filepath.Join(sessionDir, "session.json"),
	}

	// Handler of FLOOD_WAIT that will automatically retry request.
	waiter := floodwait.NewWaiter().WithCallback(func(ctx context.Context, wait floodwait.FloodWait) {
		logger.Debug(ctx, "Got FLOOD_WAIT. Will retry", zap.Duration("after", wait.Duration))
	})

	// Filling client options.
	options := telegram.Options{
		SessionStorage: sessionStorage, // Setting up session sessionStorage to store auth data.
		Middlewares: []telegram.Middleware{
			// Setting up FLOOD_WAIT handler to automatically wait and retry request.
			waiter,
			ratelimit.New(rate.Every(time.Second), 3),
		},
	}
	client := telegram.NewClient(appID, appHash, options)

	// Authentication flow handles authentication process, like prompting for code and 2FA password.
	flow := auth.NewFlow(examples.Terminal{PhoneNumber: phone}, auth.SendCodeOptions{})

	return waiter.Run(ctx, func(ctx context.Context) error {
		// Spawning main goroutine.
		if err := client.Run(ctx, func(ctx context.Context) error {
			// Perform auth if no session is available.
			if err := client.Auth().IfNecessary(ctx, flow); err != nil {
				return errors.Wrap(err, "auth")
			}

			// // Getting info about current user.
			// self, err := client.Self(ctx)
			// if err != nil {
			// 	return errors.Wrap(err, "call self")
			// }

			// name := self.FirstName
			// if self.Username != "" {
			// 	// Username is optional.
			// 	name = fmt.Sprintf("%s (@%s)", name, self.Username)
			// }
			// fmt.Println("Current user:", name)

			storage := postgres.New(postgresConfig)
			if err := storage.Ready(ctx); err != nil {
				return fmt.Errorf("storage not ready: %v", err)
			}
			defer storage.Close()

			totalMessages := 0
			for i, chatIDSecret := range cfg.ChatIDs {
				chatID := chatIDSecret.Unmask()
				logger.Debug(ctx, "Processing chat", zap.Int("current", i+1), zap.Int("total", len(cfg.ChatIDs)), zap.Int64("chat_id", chatID))

				messages, err := readMessagesForChat(ctx, client, storage, chatID)
				if err != nil {
					return fmt.Errorf("error reading messages from chat %d: %v", chatID, err)
				}

				totalMessages += messages
				logger.Debug(ctx, "Chat processed", zap.Int64("chat_id", chatID), zap.Int("messages", messages))
			}

			logger.Debug(ctx, "Total messages processed", zap.Int("total", totalMessages))
			return nil
		}); err != nil {
			return errors.Wrap(err, "run")
		}
		return nil
	})
}

func processMessages(messagesList []tg.MessageClass) []*tg.Message {
	var result []*tg.Message

	for _, msg := range messagesList {
		if message, ok := msg.(*tg.Message); ok {
			result = append(result, message)
		}
	}

	return result
}

func getMessagesList(response interface{}) ([]tg.MessageClass, error) {
	switch resp := response.(type) {
	case *tg.MessagesMessages:
		return resp.Messages, nil
	case *tg.MessagesMessagesSlice:
		return resp.Messages, nil
	case *tg.MessagesChannelMessages:
		return resp.Messages, nil
	default:
		return nil, fmt.Errorf("unexpected response type: %T", response)
	}
}

func readMessagesFromChannel(ctx context.Context, client *telegram.Client, storage *postgres.Storage, channelID int64, accessHash int64, chatIdentifier string) (int, error) {
	totalProcessed := 0
	var wg sync.WaitGroup
	var lastProcessedID int64

	offsetID := 0
	if _, err := storage.CreateChat(ctx, storagemodels.ReqCreateChat{
		ChatID: storagemodels.ChatID(chatIdentifier),
	}); err != nil {
		return totalProcessed, fmt.Errorf("failed to create chat %s: %w", chatIdentifier, err)
	}
	lastReadResp, err := storage.GetChatLastReadID(ctx, storagemodels.ReqGetChatLastReadID{
		ChatID: chatIdentifier,
	})
	if err != nil {
		return totalProcessed, fmt.Errorf("failed to get last read ID for chat %s: %w", chatIdentifier, err)
	}
	startID := int(lastReadResp.LastReadID)
	peer := &tg.InputPeerChannel{
		ChannelID:  channelID,
		AccessHash: accessHash,
	}
	for {
		resp, err := client.API().MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
			Peer:     peer,
			OffsetID: offsetID,
			Limit:    100,
			MinID:    startID,
		})
		if err != nil {
			return totalProcessed, err
		}

		messagesList, err := getMessagesList(resp)
		if err != nil {
			return totalProcessed, err
		}

		processed := processMessages(messagesList)
		if len(processed) == 0 {
			break
		}

		if len(processed) > 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := writeMessagesToDatabase(context.Background(), storage, processed, chatIdentifier)
				if err != nil {
					logger.Error(context.Background(), fmt.Errorf("error writing batch to database: %w", err))
				}
			}()
			totalProcessed += len(processed)
		}

		logger.Debug(ctx, "Messages processed", zap.Int("count", len(processed)), zap.Int("last_id", processed[len(processed)-1].ID), zap.Int("saved", len(processed)))

		if len(processed) > 0 {
			firstMessageID := int64(processed[0].ID)
			if firstMessageID > lastProcessedID {
				lastProcessedID = firstMessageID
			}
		}
		offsetID = int(processed[len(processed)-1].ID)
	}

	wg.Wait()

	if lastProcessedID > 0 {
		_, err := storage.UpdateChatLastReadID(ctx, storagemodels.ReqUpdateChatLastReadID{
			ChatID:     chatIdentifier,
			LastReadID: lastProcessedID,
		})
		if err != nil {
			return totalProcessed, fmt.Errorf("failed to update last read ID for chat %s: %w", chatIdentifier, err)
		}
	}

	return totalProcessed, nil
}

func readMessagesForChat(ctx context.Context, client *telegram.Client, storage *postgres.Storage, chatID int64) (int, error) {
	dialogResp, err := client.API().MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
	})
	if err != nil {
		return 0, fmt.Errorf("error getting dialogs: %v", err)
	}

	modified, ok := dialogResp.AsModified()
	if !ok {
		return 0, fmt.Errorf("unexpected dialogs response type: %T", dialogResp)
	}

	chats := modified.GetChats()
	chatIdentifier := fmt.Sprintf("%d", chatID)

	for _, dialog := range modified.GetDialogs() {
		switch peer := dialog.GetPeer().(type) {
		case *tg.PeerChat:
			if peer.ChatID == chatID {
				return readMessagesFromChannel(ctx, client, storage, chatID, 0, chatIdentifier)
			}
		case *tg.PeerChannel:
			if peer.ChannelID == chatID {
				for _, chatClass := range chats {
					if ch, ok := chatClass.(*tg.Channel); ok && ch.ID == chatID {
						return readMessagesFromChannel(ctx, client, storage, chatID, ch.AccessHash, chatIdentifier)
					}
				}
			}
		}
	}

	return 0, fmt.Errorf("chat with ID %d not found in dialogs", chatID)
}
