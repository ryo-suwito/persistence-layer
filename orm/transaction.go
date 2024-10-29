package orm

import (
    "persistence-layer/adapters"
)

// Transaction interface defines the operations for a SQL transaction.
type Transaction interface {
    Commit() error
    Rollback() error
    Create(model interface{}) error
    Update(model interface{}) error
    Delete(id uint, model interface{}) error
}

// SQLTransaction implements the Transaction interface using a SQL adapter.
type SQLTransaction struct {
    tx *adapters.SQLAdapter
}

// NewSQLTransaction creates a new SQLTransaction using the provided adapter.
func NewSQLTransaction(adapter *adapters.SQLAdapter) (*SQLTransaction, error) {
    tx, err := adapter.BeginTransaction()
    if err != nil {
        return nil, err
    }
    return &SQLTransaction{tx: tx}, nil
}

// Commit commits the transaction.
func (t *SQLTransaction) Commit() error {
    return t.tx.Commit()
}

// Rollback rolls back the transaction.
func (t *SQLTransaction) Rollback() error {
    return t.tx.Rollback()
}

// Create inserts a new record within the transaction.
func (t *SQLTransaction) Create(model interface{}) error {
    return t.tx.Create(model)
}

// Update updates an existing record within the transaction.
func (t *SQLTransaction) Update(model interface{}) error {
    return t.tx.Update(model)
}

// Delete removes a record within the transaction.
func (t *SQLTransaction) Delete(id uint, model interface{}) error {
    return t.tx.Delete(id, model)
}
