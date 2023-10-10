package brows

import (
	"context"
	"database/sql"
)

type Query interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type Brows struct {
	query Query
}

// New return new Brows
//
// query could be *sql.DB, *sql.Tx, *sql.Conn or other object who implemented Query interface
func New(query Query) *Brows {
	return &Brows{
		query: query,
	}
}

func (b *Brows) QueryRow(query string, args ...any) *Row {
	return b.QueryRowContext(context.Background(), query, args...)
}

func (b *Brows) QueryRowContext(ctx context.Context, query string, args ...any) *Row {
	rows, err := b.query.QueryContext(ctx, query, args...)
	return &Row{err: err, rows: rows}
}

type Row struct {
	err  error
	rows *sql.Rows
}

func (r *Row) Err() error {
	return r.err
}

func (r *Row) Scan(dest any) error {
	if r.err != nil {
		return r.err
	}
	return Scan(r.rows, dest)
}

func (b *Brows) Query(query string, args ...any) *Rows {
	return b.QueryContext(context.Background(), query, args...)
}

func (b *Brows) QueryContext(ctx context.Context, query string, args ...any) *Rows {
	rows, err := b.query.QueryContext(ctx, query, args...)
	return &Rows{err: err, rows: rows}
}

type Rows struct {
	err  error
	rows *sql.Rows
}

func (rs *Rows) Err() error {
	return rs.err
}

func (rs *Rows) Scan(dest any) error {
	if rs.err != nil {
		return rs.err
	}
	return ScanSlice(rs.rows, dest)
}
