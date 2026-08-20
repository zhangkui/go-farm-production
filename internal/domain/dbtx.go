package domain

import (
	"context"
	"database/sql"
)

// DBTX is the common interface between *sql.DB and *sql.Tx so that repository
// code can run either against a plain connection pool or inside a transaction.
// Using the standard library interface keeps dependencies minimal.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// TxStarter extends DBTX with the ability to begin transactions, which is only
// available on *sql.DB. Repository Store implementations satisfy this.
type TxStarter interface {
	DBTX
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// TxRunner is implemented by the Store and lets services execute a unit of work
// against a transactional DBTX. The callback receives a DBTX bound to the tx.
type TxRunner interface {
	WithTx(ctx context.Context, fn func(ctx context.Context, tx DBTX) error) error
}

// RowsScanner closes rows and scans into a slice of struct pointers via a
// caller-provided scan function. Centralises the rows-loop boilerplate.
func ScanRows[T any](rows *sql.Rows, scan func(*sql.Rows) (T, error)) ([]T, error) {
	defer rows.Close()
	out := make([]T, 0)
	for rows.Next() {
		v, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
