package wrapper

import (
	"context"
	"database/sql/driver"

	"github.com/yanakipre/bot/internal/rdb/internal/driver/sqldata"
)

// implements driver.Stmt
type wrStmt struct {
	// name
	// statement name
	name    string
	parent  driver.Stmt
	query   string
	wrapper Hook
}

func valuesToInterfaces(args []driver.Value) []any {
	res := make([]any, 0, len(args))
	for _, arg := range args {
		res = append(res, arg)
	}
	return res
}

func namedValuesToInterfaces(args []driver.NamedValue) []any {
	// TODO: fix this - we got the named values
	res := make([]any, 0, len(args))
	for _, arg := range args {
		res = append(res, arg)
	}
	return res
}

func (s wrStmt) Exec(args []driver.Value) (res driver.Result, err error) {
	list := valuesToInterfaces(args)
	ctx := sqldata.NewContext(
		context.Background(),
		sqldata.Stmt(s.query, list...),
		sqldata.Action(sqldata.ActionExec),
		sqldata.Operation(s.name),
	)
	ctx = s.wrapper.Before(ctx)
	res, err = s.parent.Exec(args) //nolint:staticcheck
	s.wrapper.After(ctx, err)
	if err != nil {
		return nil, err
	}
	return
}

func (s wrStmt) Close() error {
	return s.parent.Close()
}

func (s wrStmt) NumInput() int {
	return s.parent.NumInput()
}

func (s wrStmt) Query(args []driver.Value) (rows driver.Rows, err error) {
	list := valuesToInterfaces(args)
	ctx := sqldata.NewContext(
		context.Background(),
		sqldata.Stmt(s.query, list...),
		sqldata.Action(sqldata.ActionQuery),
		sqldata.Operation(s.name),
	)
	ctx = s.wrapper.Before(ctx)
	rows, err = s.parent.Query(args) //nolint:staticcheck
	s.wrapper.After(ctx, err)
	if err != nil {
		return nil, err
	}
	return
}

func (s wrStmt) ExecContext(
	ctx context.Context,
	args []driver.NamedValue,
) (res driver.Result, err error) {
	// we already tested driver to implement StmtExecContext
	execContext := s.parent.(driver.StmtExecContext)
	list := namedValuesToInterfaces(args)
	ctx = sqldata.NewContext(
		ctx,
		sqldata.Stmt(s.query, list...),
		sqldata.Action(sqldata.ActionExec),
		sqldata.Operation(s.name),
	)
	ctx = s.wrapper.Before(ctx)
	res, err = execContext.ExecContext(ctx, args)
	s.wrapper.After(ctx, err)
	if err != nil {
		return nil, err
	}
	return
}

func (s wrStmt) QueryContext(
	ctx context.Context,
	args []driver.NamedValue,
) (rows driver.Rows, err error) {
	// we already tested driver to implement StmtQueryContext
	queryContext := s.parent.(driver.StmtQueryContext)
	list := namedValuesToInterfaces(args)
	ctx = sqldata.NewContext(
		ctx,
		sqldata.Stmt(s.query, list...),
		sqldata.Action(sqldata.ActionQuery),
		sqldata.Operation(s.name),
	)
	ctx = s.wrapper.Before(ctx)
	rows, err = queryContext.QueryContext(ctx, args)
	s.wrapper.After(ctx, err)
	if err != nil {
		return nil, err
	}
	return
}
