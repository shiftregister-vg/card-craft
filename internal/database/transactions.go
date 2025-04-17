package database

import (
	"context"
	"database/sql"
	"errors"
)

// Transaction represents a database transaction
type Transaction struct {
	tx *sql.Tx
}

// Begin starts a new transaction
func Begin(ctx context.Context, db *sql.DB) (*Transaction, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &Transaction{tx: tx}, nil
}

// Commit commits the transaction
func (t *Transaction) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback() error {
	return t.tx.Rollback()
}

// Exec executes a query within the transaction
func (t *Transaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	return t.tx.Exec(query, args...)
}

// Query executes a query that returns rows within the transaction
func (t *Transaction) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.Query(query, args...)
}

// QueryRow executes a query that is expected to return at most one row within the transaction
func (t *Transaction) QueryRow(query string, args ...interface{}) *sql.Row {
	return t.tx.QueryRow(query, args...)
}

// WithTransaction executes a function within a transaction
func WithTransaction(ctx context.Context, db *sql.DB, fn func(*Transaction) error) error {
	tx, err := Begin(ctx, db)
	if err != nil {
		return err
	}

	// Ensure we rollback if there's an error
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	// Execute the function
	err = fn(tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// Commit the transaction
	return tx.Commit()
}

// TransactionError wraps multiple errors that occurred during a transaction
type TransactionError struct {
	Errors []error
}

func (e *TransactionError) Error() string {
	return "multiple errors occurred during transaction"
}

// IsTransactionError checks if an error is a TransactionError
func IsTransactionError(err error) bool {
	var te *TransactionError
	return errors.As(err, &te)
}
