package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Reader interface {
	ExecContext(ctx context.Context, query string, args pgx.QueryRewriter) (count int64, err error)
}

type Writer[T any] interface {
	QueryRowContext(ctx context.Context, scanner func(row pgx.Row, t *T) error, query string, args pgx.QueryRewriter) (*T, error)
	QueryContext(ctx context.Context, scanner func(row pgx.Rows, t *T) error, query string, args pgx.QueryRewriter) ([]*T, error)
}

type Conn[T any] interface {
	Conn() *pgxpool.Pool

	Reader
	Writer[T]
}

type connHandler[T any] struct {
	conn *pgxpool.Pool
}

func NewConn[T any](conn *pgxpool.Pool) Conn[T] {
	return &connHandler[T]{conn: conn}
}

func (h *connHandler[T]) Conn() *pgxpool.Pool {
	return h.conn
}

func (h *connHandler[T]) ExecContext(ctx context.Context, query string, args pgx.QueryRewriter) (int64, error) {
	return ExecContext(ctx, h.conn, query, args)
}

func (h *connHandler[T]) QueryRowContext(ctx context.Context, scanner func(row pgx.Row, t *T) error, query string, args pgx.QueryRewriter) (*T, error) {
	return QueryRowContext(ctx, h.conn, scanner, query, args)
}

func (h *connHandler[T]) QueryContext(ctx context.Context, scanner func(row pgx.Rows, t *T) error, query string, args pgx.QueryRewriter) ([]*T, error) {
	return QueryContext(ctx, h.conn, scanner, query, args)
}

func ExecContext(ctx context.Context, q *pgxpool.Pool, query string, args ...any) (int64, error) {
	tag, err := q.Exec(ctx, query, args...)
	return tag.RowsAffected(), err
}

func QueryRowContext[T any](ctx context.Context, q *pgxpool.Pool, scanner func(r pgx.Row, t *T) error, query string, args ...any) (*T, error) {
	var t T
	err := scanner(q.QueryRow(ctx, query, args...), &t)
	return &t, err
}

func QueryContext[T any](ctx context.Context, q *pgxpool.Pool, scanner func(r pgx.Rows, v *T) error, query string, args ...any) ([]*T, error) {
	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var vs []*T
	for rows.Next() {
		var v T
		err = scanner(rows, &v)
		if err != nil {
			return nil, err
		}
		vs = append(vs, &v)
	}
	return vs, rows.Err()
}

var ErrNoRowsAffected = errors.New("no rows affected in result set")
