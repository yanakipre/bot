package appv1

import (
	"context"

	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

type App struct {
}

func New(d Deps) *App {
	ctx := context.TODO()
	err := d.client.Ping(ctx)
	if err != nil {
		d.lg.Fatal("failed to ping", zap.Error(err))
	}
	//bld := query.NewQuery(d.client).Messages().GetHistory(&tg.InputPeerChat{
	//	ChatID: 0, // TODO: chat id
	//})
	//bld.Iter().Value().Entities.
	history, err := d.client.API().MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
		Peer: &tg.InputPeerChat{
			ChatID: 0, // TODO: chat id
		},
		Limit:    100,
		OffsetID: 100,
	})
	switch a := history.(type) {
	case *tg.MessagesMessages:
		if len(a.Messages) == 0 {
			d.lg.Fatal("no messages")
		}
		switch b := a.Messages[0].(type) {
		case *tg.Message:
			d.lg.Info("message", zap.String("message", b.Message))
		default:
			d.lg.Fatal("unexpected msg type")
		}
	default:
		d.lg.Fatal("unexpected msgs type")
	}
	if err != nil {
		d.lg.Fatal("failed to get history", zap.Error(err))
	}
	return nil

}
