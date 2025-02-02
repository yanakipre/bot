package wrapper

import (
	"context"
	"database/sql/driver"
)

// Compile time assertion
var (
	_ driver.DriverContext = &wrDriver{}
	_ driver.Connector     = &wrDriver{}
)

// WrapConnector allows wrapping a database driver.Connector which eliminates
// the need to register wrapped sql as an available driver.Driver.
func WrapConnector(dc driver.Connector) driver.Connector {
	return &wrDriver{
		parent:    dc.Driver(),
		connector: dc,
	}
}

type AfterConnQuery func(ctx context.Context, conn driver.Conn) error

func dummyAfterConnQuery(ctx context.Context, conn driver.Conn) error {
	return nil
}

// wrDriver implements driver.Driver
type wrDriver struct {
	connHook  AfterConnQuery
	parent    driver.Driver
	connector driver.Connector
	wrapper   Hook
}

func wrapDriver(d driver.Driver, w Hook, conn AfterConnQuery) driver.Driver {
	if conn == nil {
		conn = dummyAfterConnQuery
	}
	if _, ok := d.(driver.DriverContext); ok {
		return wrDriver{connHook: conn, parent: d, wrapper: w}
	}
	return struct{ driver.Driver }{wrDriver{connHook: conn, parent: d, wrapper: w}}
}

func wrapStmt(stmt driver.Stmt, query string, name string, wrapper Hook) driver.Stmt {
	var (
		_, hasExeCtx    = stmt.(driver.StmtExecContext)
		_, hasQryCtx    = stmt.(driver.StmtQueryContext)
		c, hasColConv   = stmt.(driver.ColumnConverter) //nolint:staticcheck
		n, hasNamValChk = stmt.(driver.NamedValueChecker)
	)

	s := wrStmt{name: name, parent: stmt, query: query, wrapper: wrapper}
	switch {
	case !hasExeCtx && !hasQryCtx && !hasColConv && !hasNamValChk:
		return struct {
			driver.Stmt
		}{s}
	case !hasExeCtx && hasQryCtx && !hasColConv && !hasNamValChk:
		return struct {
			driver.Stmt
			driver.StmtQueryContext
		}{s, s}
	case hasExeCtx && !hasQryCtx && !hasColConv && !hasNamValChk:
		return struct {
			driver.Stmt
			driver.StmtExecContext
		}{s, s}
	case hasExeCtx && hasQryCtx && !hasColConv && !hasNamValChk:
		return struct {
			driver.Stmt
			driver.StmtExecContext
			driver.StmtQueryContext
		}{s, s, s}
	case !hasExeCtx && !hasQryCtx && hasColConv && !hasNamValChk:
		return struct {
			driver.Stmt
			driver.ColumnConverter
		}{s, c}
	case !hasExeCtx && hasQryCtx && hasColConv && !hasNamValChk:
		return struct {
			driver.Stmt
			driver.StmtQueryContext
			driver.ColumnConverter
		}{s, s, c}
	case hasExeCtx && !hasQryCtx && hasColConv && !hasNamValChk:
		return struct {
			driver.Stmt
			driver.StmtExecContext
			driver.ColumnConverter
		}{s, s, c}
	case hasExeCtx && hasQryCtx && hasColConv && !hasNamValChk:
		return struct {
			driver.Stmt
			driver.StmtExecContext
			driver.StmtQueryContext
			driver.ColumnConverter
		}{s, s, s, c}

	case !hasExeCtx && !hasQryCtx && !hasColConv && hasNamValChk:
		return struct {
			driver.Stmt
			driver.NamedValueChecker
		}{s, n}
	case !hasExeCtx && hasQryCtx && !hasColConv && hasNamValChk:
		return struct {
			driver.Stmt
			driver.StmtQueryContext
			driver.NamedValueChecker
		}{s, s, n}
	case hasExeCtx && !hasQryCtx && !hasColConv && hasNamValChk:
		return struct {
			driver.Stmt
			driver.StmtExecContext
			driver.NamedValueChecker
		}{s, s, n}
	case hasExeCtx && hasQryCtx && !hasColConv && hasNamValChk:
		return struct {
			driver.Stmt
			driver.StmtExecContext
			driver.StmtQueryContext
			driver.NamedValueChecker
		}{s, s, s, n}
	case !hasExeCtx && !hasQryCtx && hasColConv && hasNamValChk:
		return struct {
			driver.Stmt
			driver.ColumnConverter
			driver.NamedValueChecker
		}{s, c, n}
	case !hasExeCtx && hasQryCtx && hasColConv && hasNamValChk:
		return struct {
			driver.Stmt
			driver.StmtQueryContext
			driver.ColumnConverter
			driver.NamedValueChecker
		}{s, s, c, n}
	case hasExeCtx && !hasQryCtx && hasColConv && hasNamValChk:
		return struct {
			driver.Stmt
			driver.StmtExecContext
			driver.ColumnConverter
			driver.NamedValueChecker
		}{s, s, c, n}
	case hasExeCtx && hasQryCtx && hasColConv && hasNamValChk:
		return struct {
			driver.Stmt
			driver.StmtExecContext
			driver.StmtQueryContext
			driver.ColumnConverter
			driver.NamedValueChecker
		}{s, s, s, c, n}
	}
	panic("unreachable")
}

func (d wrDriver) OpenConnector(name string) (driver.Connector, error) {
	var err error
	d.connector, err = d.parent.(driver.DriverContext).OpenConnector(name)
	if err != nil {
		return nil, err
	}
	return d, err
}

func (d wrDriver) Connect(ctx context.Context) (driver.Conn, error) {
	c, err := d.connector.Connect(ctx)
	if err != nil {
		return nil, err
	}
	result := &wrConn{parent: c, wrapper: connIDHook{genConnID(), d.wrapper}}
	if err := d.connHook(ctx, result); err != nil {
		return nil, err
	}

	return result, nil
}

func (d wrDriver) Driver() driver.Driver {
	return d
}
