package wrapper

import (
	"context"
	"database/sql/driver"

	"github.com/yanakipe/bot/internal/rdb/internal/driver/sqldata"
)

// wrTx implements driver.Tx
type wrTx struct {
	parent  driver.Tx
	ctx     context.Context
	txctx   context.Context
	wrapper Hook
}

func (t wrTx) Commit() (err error) {
	ctx := sqldata.NewContext(t.ctx, sqldata.Action(sqldata.ActionCommit))
	ctx = t.wrapper.Before(ctx)
	err = t.parent.Commit()
	t.wrapper.After(ctx, err)
	t.wrapper.After(t.txctx, err)
	return
}

func (t wrTx) Rollback() (err error) {
	ctx := sqldata.NewContext(t.ctx, sqldata.Action(sqldata.ActionRollback))
	ctx = t.wrapper.Before(ctx)
	err = t.parent.Rollback()
	t.wrapper.After(ctx, err)
	t.wrapper.After(t.txctx, err)
	return
}
