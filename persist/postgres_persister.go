package persist

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"github.com/lib/pq"
	"github.com/ourrootsorg/cms-server/model"
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

func translateError(err error, id *uint32, refID *uint32, refType string) error {
	switch err {
	case nil:
		return nil
	case sql.ErrConnDone:
		return model.NewError(model.ErrOther, err.Error())
	case sql.ErrNoRows:
		var sid string
		if id != nil {
			sid = strconv.Itoa(int(*id))
		}
		return model.NewError(model.ErrNotFound, sid)
	case sql.ErrTxDone:
		return model.NewError(model.ErrOther, err.Error())
	default:
		pqErr, ok := err.(*pq.Error)
		if !ok {
			log.Printf("[INFO] Untranslated error: %#v", err)
			return model.NewError(model.ErrOther, err.Error())
		}
		switch pqErr.Code.Name() {
		case "foreign_key_violation":
			if refID != nil {
				return model.NewError(model.ErrBadReference, strconv.Itoa(int(*refID)), refType)
			}
			return model.NewError(model.ErrOther, fmt.Sprintf("Foreign key violation, but reference information not supplied: %v", err))
		default:
			log.Printf("[INFO] Untranslated PQ error: %#v", err)
			return model.NewError(model.ErrOther, err.Error())
		}
	}
}
