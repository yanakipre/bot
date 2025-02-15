package sqldata

import (
	"context"
)

const (
	ActionPrepare  ActionEnum = "prepare"
	ActionQuery    ActionEnum = "query"
	ActionExec     ActionEnum = "exec"
	ActionTx       ActionEnum = "tx"
	ActionBegin    ActionEnum = "begin"
	ActionCommit   ActionEnum = "commit"
	ActionRollback ActionEnum = "rollback"
)

// NewContext gets previous sqldata.Data from context, sets new opts for it and stores it in context
func NewContext(ctx context.Context, opts ...Option) context.Context {
	data := FromContext(ctx)
	for _, o := range opts {
		o(&data)
	}
	return context.WithValue(ctx, DataKey, data)
}

func FromContext(ctx context.Context) Data {
	d, _ := ctx.Value(DataKey).(Data)
	return d
}

// Operation gives short description of the operation which is executed
func Operation(o string) Option {
	return func(d *Data) {
		d.Operation = o
	}
}

func Action(action ActionEnum) Option {
	return func(d *Data) {
		d.Action = action
	}
}

func Stmt(stmt string, args ...any) Option {
	return func(d *Data) {
		d.Stmt = stmt
		d.Args = args
	}
}
