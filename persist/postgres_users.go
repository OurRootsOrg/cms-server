package persist

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/ourrootsorg/cms-server/model"
)

// User persistence mthods

// RetrieveUser either retrieves a user record from the database, or creates the record if it doesn't
// already exist.
func (p PostgresPersister) RetrieveUser(ctx context.Context, in model.UserIn) (*model.User, error) {
	var user model.User
	var dbID int32
	log.Printf("[DEBUG] Looking up subject '%s' in database", in.Subject)
	err := p.db.QueryRowContext(ctx, `SELECT id, body, insert_time, last_update_time 
		FROM cms_user
		WHERE body->>'iss'=$1 AND body->>'sub'=$2`, in.Issuer, in.Subject).
		Scan(
			&dbID,
			&user.UserBody,
			&user.InsertTime,
			&user.LastUpdateTime,
		)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("[DEBUG] Error looking up subject '%s' in database", in.Subject)
		return nil, translateError(err)
	}
	if err == nil {
		log.Printf("[DEBUG] Found subject '%s' in database", in.Subject)
		user.ID = model.MakeUserID(dbID)
		if !user.Enabled {
			msg := fmt.Sprintf("User '%s' is not enabled", user.ID)
			log.Printf("[DEBUG] %s", msg)
			return nil, errors.New(msg)
		}
		// We got a user
		log.Printf("[DEBUG] Returning enabled user '%s'", user.ID)
		return &user, nil
	}
	log.Printf("[DEBUG] No user with subject '%s' found in database, so creating one", in.Subject)
	err = p.db.QueryRowContext(ctx,
		`INSERT INTO cms_user (body) 
		 VALUES ($1) 
		 RETURNING id, body, insert_time, last_update_time`,
		in.UserBody).
		Scan(
			&dbID,
			&user.UserBody,
			&user.InsertTime,
			&user.LastUpdateTime,
		)
	user.ID = model.MakeUserID(dbID)
	log.Printf("[DEBUG] Created user '%s'", user.ID)
	return &user, translateError(err)
}
