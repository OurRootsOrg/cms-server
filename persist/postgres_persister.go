package persist

import (
	"database/sql"
	"log"

	"github.com/lib/pq"
)

// PostgresPersister persists the model objects to Postgresql
type PostgresPersister struct {
	db *sql.DB
}

// NewPostgresPersister constructs a new PostgresPersister
func NewPostgresPersister(db *sql.DB) PostgresPersister {
	return PostgresPersister{
		db: db,
	}
}

func translateError(err error) error {
	switch err {
	case nil:
		return nil
	case sql.ErrConnDone:
		return ErrConnDone
	case sql.ErrNoRows:
		return ErrNoRows
	case sql.ErrTxDone:
		return ErrTxDone
	default:
		pqErr, ok := err.(*pq.Error)
		if !ok {
			log.Printf("[INFO] Untranslated error: %#v", err)
			return err
		}
		switch pqErr.Code.Name() {
		case "foreign_key_violation":
			return ErrForeignKeyViolation
		default:
			log.Printf("[INFO] Untranslated PQ error: %#v", err)
			return err
		}
	}
}
