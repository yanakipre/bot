// Package rdb when handling the errors for the database operations, has two paths:
// 1. When outside the transaction, retryable errors are OK to be simply retries
// 2. When inside the transaction, any retryable error should lead to retrying the whole transaction.
package rdb

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/kamilsk/retry/v5"
)

type sqlxdbWrapper struct {
	*sqlx.DB
	retries retry.How
}

func (o sqlxdbWrapper) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	var r *sql.Rows
	s := o.retries
	err := retry.Do(ctx, func(ctx context.Context) error {
		queryContext, err := o.DB.QueryContext(ctx, query, args...)
		r = queryContext
		return err
	}, s...)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (o sqlxdbWrapper) QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error) {
	var r *sqlx.Rows
	s := o.retries
	err := retry.Do(ctx, func(ctx context.Context) error {
		queryContext, err := o.DB.QueryxContext(ctx, query, args...)
		r = queryContext
		return err
	}, s...)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// TODO: should be changed. No error can be wrapped here.
func (o sqlxdbWrapper) QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row {
	return o.DB.QueryRowxContext(ctx, query, args...)
}

func (o sqlxdbWrapper) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	var r sql.Result
	s := o.retries
	err := retry.Do(ctx, func(ctx context.Context) error {
		execContext, err := o.DB.ExecContext(ctx, query, args...)
		r = execContext
		return err
	}, s...)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (o sqlxdbWrapper) PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error) {
	var r *sqlx.NamedStmt
	s := o.retries
	err := retry.Do(ctx, func(ctx context.Context) error {
		prepareNamedContext, err := o.DB.PrepareNamedContext(ctx, query)
		r = prepareNamedContext
		return err
	}, s...)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (o sqlxdbWrapper) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	s := o.retries
	return retry.Do(ctx, func(ctx context.Context) error {
		return o.DB.SelectContext(ctx, dest, query, args...)
	}, s...)
}

func (o sqlxdbWrapper) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	s := o.retries
	return retry.Do(ctx, func(ctx context.Context) error {
		return o.DB.GetContext(ctx, dest, query, args...)
	}, s...)
}

var _ TxExecutor = &sqlxdbWrapper{}
