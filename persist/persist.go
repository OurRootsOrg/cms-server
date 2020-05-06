package persist

import (
	"errors"
)

// Error definitions

// ErrNoRows is returned when there are no rows
var ErrNoRows = errors.New("persist: no rows in result set")

// ErrTxDone is returned by any operation that is performed on a transaction
// that has already been committed or rolled back.
var ErrTxDone = errors.New("persist: transaction has already been committed or rolled back")

// ErrConnDone is returned by any operation that is performed on a connection
// that has already been returned to the connection pool.
var ErrConnDone = errors.New("persist: connection is already closed")

// ErrForeignKeyViolation is returned by an operation that causes a foreign key violation
var ErrForeignKeyViolation = errors.New("persist: foreign_key_violation")
