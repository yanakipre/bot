package rdb

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type contextKey string

// TxKey is used to store current running *sqlx.Tx inside context.
const TxKey = contextKey("tx")

const TxRollbackKey = contextKey("tx-rollback")

// afterCommitCallbacks stores a list of closures to execute after commit.
const afterCommitCallbacksKey = contextKey("afterCommitCallbacks")

type TxCallback func(ctx context.Context)

type TxExecutor interface {
	// ExtContext is a base interface with Query/ExecContext functions
	sqlx.ExtContext

	// additional methods
	PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error)
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
	GetContext(ctx context.Context, dest any, query string, args ...any) error
}

func IsInsideTx(ctx context.Context) bool {
	_, ok := ctx.Value(TxKey).(*sqlx.Tx)
	return ok
}

func TxRollbackFunc(ctx context.Context) func() error {
	rollback, ok := ctx.Value(TxRollbackKey).(func() error)
	if ok && rollback != nil {
		return rollback
	}
	return nil
}
