package wrapper

import (
	"context"
	"database/sql/driver"

	"github.com/yanakipre/bot/internal/rdb/internal/driver/sqldata"
)

// WrapConn allows an existing driver.Conn to be wrapped.
func WrapConn(c driver.Conn, w Hook) driver.Conn {
	return wrapConn(c, w)
}

// UnwrapConn return parent driver.Conn .
func UnwrapConn(c driver.Conn) driver.Conn {
	return unwrapConn(c)
}

func unwrapConn(c driver.Conn) driver.Conn {
	if c == nil {
		return nil
	}

	parent := c

	for {
		switch t := parent.(type) {
		case wrConnWithSessionResetter:
			parent = t.parent
		case wrConnWithNameValueChecker:
			parent = t.parent
		case wrConnWithNameValueCheckerAndWithSessionResetter:
			parent = t.parent
		default:
			return parent
		}
	}
}

type wrConnWithNameValueChecker struct {
	*wrConn
	driver.NamedValueChecker
}

type wrConnWithSessionResetter struct {
	*wrConn
	driver.SessionResetter
}

type wrConnWithNameValueCheckerAndWithSessionResetter struct {
	*wrConn
	driver.NamedValueChecker
	driver.SessionResetter
}

func wrapConn(parent driver.Conn, w Hook) driver.Conn {
	var (
		n, hasNameValueChecker = parent.(driver.NamedValueChecker)
		s, hasSessionResetter  = parent.(driver.SessionResetter)
	)
	c := &wrConn{parent: parent, wrapper: connIDHook{
		connID: genConnID(),
		hook:   w,
	}}
	switch {
	case !hasNameValueChecker && !hasSessionResetter:
		return c
	case hasNameValueChecker && !hasSessionResetter:
		return wrConnWithNameValueChecker{c, n}
	case !hasNameValueChecker && hasSessionResetter:
		return wrConnWithSessionResetter{c, s}
	case hasNameValueChecker && hasSessionResetter:
		return wrConnWithNameValueCheckerAndWithSessionResetter{c, n, s}
	}
	panic("unreachable")
}

// wrConn implements driver.Conn
type wrConn struct {
	parent  driver.Conn
	wrapper Hook
}

func (c wrConn) Ping(ctx context.Context) (err error) {
	if pinger, ok := c.parent.(driver.Pinger); ok {
		err = pinger.Ping(ctx)
	}
	return
}

func (c wrConn) Exec(query string, args []driver.Value) (res driver.Result, err error) {
	if exec, ok := c.parent.(driver.Execer); ok { //nolint:staticcheck,nolintlint
		list := valuesToInterfaces(args)
		ctx := sqldata.NewContext(
			context.Background(),
			sqldata.Stmt(query, list...),
			sqldata.Action(sqldata.ActionExec),
		)

		ctx = c.wrapper.Before(ctx)
		res, err = exec.Exec(query, args)
		c.wrapper.After(ctx, err)
		if err != nil {
			return nil, err
		}
		return
	}

	return nil, driver.ErrSkip
}

func (c wrConn) ExecContext(
	ctx context.Context,
	query string,
	args []driver.NamedValue,
) (res driver.Result, err error) {
	if execContext, ok := c.parent.(driver.ExecerContext); ok {
		list := namedValuesToInterfaces(args)
		execCtx := sqldata.NewContext(
			ctx,
			sqldata.Stmt(query, list...),
			sqldata.Action(sqldata.ActionExec),
		)

		execCtx = c.wrapper.Before(execCtx)
		res, err = execContext.ExecContext(execCtx, query, args)
		c.wrapper.After(execCtx, err)
		if err != nil {
			return nil, err
		}
		return
	}

	return nil, driver.ErrSkip
}

func (c wrConn) Query(query string, args []driver.Value) (rows driver.Rows, err error) {
	if queryer, ok := c.parent.(driver.Queryer); ok { //nolint:staticcheck,nolintlint
		list := valuesToInterfaces(args)
		ctx := sqldata.NewContext(
			context.Background(),
			sqldata.Stmt(query, list...),
			sqldata.Action(sqldata.ActionQuery),
		)
		ctx = c.wrapper.Before(ctx)
		rows, err = queryer.Query(query, args)
		c.wrapper.After(ctx, err)
		if err != nil {
			return nil, err
		}

		return rows, nil
	}

	return nil, driver.ErrSkip
}

func (c wrConn) QueryContext(
	qCtx context.Context,
	query string,
	args []driver.NamedValue,
) (rows driver.Rows, err error) {
	if queryerContext, ok := c.parent.(driver.QueryerContext); ok {
		list := namedValuesToInterfaces(args)
		qCtx = sqldata.NewContext(
			qCtx,
			sqldata.Stmt(query, list...),
			sqldata.Action(sqldata.ActionQuery),
		)
		qCtx = c.wrapper.Before(qCtx)
		rows, err = queryerContext.QueryContext(qCtx, query, args)
		c.wrapper.After(qCtx, err)
		if err != nil {
			return nil, err
		}

		return rows, nil
	}

	return nil, driver.ErrSkip
}

const queryNameNotDerivedFromContext = "query name is not derived from context"

func (c wrConn) Prepare(query string) (stmt driver.Stmt, err error) {
	ctx := sqldata.NewContext(
		context.Background(),
		sqldata.Stmt(query),
		sqldata.Action(sqldata.ActionPrepare),
	)
	ctx = c.wrapper.Before(ctx)
	stmt, err = c.parent.Prepare(query)
	c.wrapper.After(ctx, err)
	if err != nil {
		return nil, err
	}

	stmt = wrapStmt(stmt, query, queryNameNotDerivedFromContext, c.wrapper)
	return
}

func (c *wrConn) Close() error {
	return c.parent.Close()
}

func (c *wrConn) PrepareContext(ctx context.Context, query string) (stmt driver.Stmt, err error) {
	ctx = sqldata.NewContext(ctx, sqldata.Stmt(query), sqldata.Action(sqldata.ActionPrepare))
	ctx = c.wrapper.Before(ctx)
	if prepCtx, ok := c.parent.(driver.ConnPrepareContext); ok {
		stmt, err = prepCtx.PrepareContext(ctx, query)
	} else {
		stmt, err = c.parent.Prepare(query)
	}
	c.wrapper.After(ctx, err)
	if err != nil {
		return nil, err
	}

	qName := queryNameNotDerivedFromContext
	if data := sqldata.FromContext(ctx); data.Operation != "" {
		qName = data.Operation
	}

	stmt = wrapStmt(stmt, query, qName, c.wrapper)
	return
}

func (c *wrConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.TODO(), driver.TxOptions{})
}

func (c *wrConn) BeginTx(ctx context.Context, opts driver.TxOptions) (tx driver.Tx, err error) {
	txctx := sqldata.NewContext(ctx, sqldata.Action(sqldata.ActionTx))
	txctx = c.wrapper.Before(txctx)
	beginctx := sqldata.NewContext(ctx, sqldata.Action(sqldata.ActionBegin))
	beginctx = c.wrapper.Before(beginctx)

	if connBeginTx, ok := c.parent.(driver.ConnBeginTx); ok {
		tx, err = connBeginTx.BeginTx(txctx, opts)
		c.wrapper.After(beginctx, err)
		if err != nil {
			c.wrapper.After(txctx, err)
			return nil, err
		}
		return wrTx{parent: tx, ctx: ctx, txctx: txctx, wrapper: c.wrapper}, nil
	}
	tx, err = c.parent.Begin() //nolint:staticcheck,nolintlint
	c.wrapper.After(beginctx, err)
	if err != nil {
		c.wrapper.After(txctx, err)
		return nil, err
	}
	return wrTx{parent: tx, ctx: ctx, txctx: txctx, wrapper: c.wrapper}, nil
}

func (c *wrConn) CheckNamedValue(nv *driver.NamedValue) (err error) {
	nvc, ok := c.parent.(driver.NamedValueChecker)
	if ok {
		return nvc.CheckNamedValue(nv)
	}
	nv.Value, err = driver.DefaultParameterConverter.ConvertValue(nv.Value)
	return err
}
