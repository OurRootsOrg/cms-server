package persist

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/ourrootsorg/cms-server/model"
)

// User persistence mthods

// RetrieveUser either retrieves a user record from the database, or creates the record if it doesn't
// already exist.
func (p PostgresPersister) RetrieveUser(ctx context.Context, in model.UserIn) (*model.User, *model.Error) {
	var user model.User
	log.Printf("[DEBUG] Looking up subject '%s' in database", in.Subject)
	err := p.db.QueryRowContext(ctx, `SELECT id, body, insert_time, last_update_time
		FROM cms_user
		WHERE body->>'iss'=$1 AND body->>'sub'=$2`, in.Issuer, in.Subject).
		Scan(
			&user.ID,
			&user.UserBody,
			&user.InsertTime,
			&user.LastUpdateTime,
		)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("[DEBUG] Error looking up subject '%s' in database", in.Subject)
		return nil, translateError(err, nil, nil, "")
	}
	if err == nil {
		log.Printf("[DEBUG] Found subject '%s' in database", in.Subject)
		if !user.Enabled {
			msg := fmt.Sprintf("User '%d' is not enabled", user.ID)
			log.Printf("[DEBUG] %s", msg)
			return nil, model.NewError(model.ErrOther, msg)
		}
		// We got a user
		log.Printf("[DEBUG] Returning enabled user '%#v'", user)
		return &user, nil
	}
	log.Printf("[DEBUG] No user with subject '%s' found in database, so creating one", in.Subject)
	err = p.db.QueryRowContext(ctx,
		`INSERT INTO cms_user (body)
		 VALUES ($1)
		 RETURNING id, body, insert_time, last_update_time`,
		in.UserBody).
		Scan(
			&user.ID,
			&user.UserBody,
			&user.InsertTime,
			&user.LastUpdateTime,
		)
	log.Printf("[DEBUG] Created user '%d'", user.ID)
	return &user, translateError(err, nil, nil, "")
}
