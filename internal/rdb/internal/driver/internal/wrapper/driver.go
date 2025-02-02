package wrapper

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"strconv"
	"sync"
)

const maxWrappedDriversCount = 1000

type conn interface {
	driver.Pinger
	driver.Execer //nolint:staticcheck
	driver.ExecerContext
	driver.Queryer //nolint:staticcheck
	driver.QueryerContext
	driver.Conn
	driver.ConnPrepareContext
	driver.ConnBeginTx
}

var (
	regMu sync.Mutex

	// Compile time assertions
	_ driver.Driver            = &wrDriver{}
	_ conn                     = &wrConn{}
	_ driver.NamedValueChecker = &wrConn{}
	_ driver.Stmt              = &wrStmt{}
	_ driver.StmtExecContext   = &wrStmt{}
	_ driver.StmtQueryContext  = &wrStmt{}
)

func WrapDriverByName(driverName string, wrapper Hook, conn AfterConnQuery) (string, error) {
	// retrieve the driver implementation we need to wrap with instrumentation
	db, err := sql.Open(driverName, "")
	if err != nil {
		return "", err
	}
	dri := db.Driver()
	if err = db.Close(); err != nil {
		return "", err
	}

	regMu.Lock()
	defer regMu.Unlock()

	// Since we might want to register multiple wrapped drivers to have different
	// TraceOptions, but potentially the same underlying database driver, we
	// cycle through to find available driver names.
	driverName = driverName + "-hooksql-"
	for i := int64(0); i < maxWrappedDriversCount; i++ {
		var (
			found   = false
			regName = driverName + strconv.FormatInt(i, 10)
		)
		for _, name := range sql.Drivers() {
			if name == regName {
				found = true
			}
		}
		if !found {
			sql.Register(regName, Wrap(dri, wrapper, conn))
			return regName, nil
		}
	}
	return "", errors.New(
		"unable to register driver, all slots have been taken." +
			" Do you want to increase `maxWrappedDriversCount`")
}

// Wrap takes a SQL driver and wraps it.
func Wrap(d driver.Driver, w Hook, conn AfterConnQuery) driver.Driver {
	return wrapDriver(d, w, conn)
}

// Open implements driver.Driver
func (d wrDriver) Open(name string) (driver.Conn, error) {
	c, err := d.parent.Open(name)
	if err != nil {
		return nil, err
	}
	return wrapConn(c, d.wrapper), nil
}
