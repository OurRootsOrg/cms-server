package persist

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/lib/pq"

	"github.com/ourrootsorg/cms-server/model"
)

// User persistence mthods

// RetrieveUser either retrieves a user record from the database, or creates the record if it doesn't
// already exist. Returns the new user and true if it was added
func (p PostgresPersister) RetrieveUser(ctx context.Context, in model.UserIn) (*model.User, bool, error) {
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
		return nil, false, translateError(err, nil, nil, "")
	}
	if err == nil {
		log.Printf("[DEBUG] Found subject '%s' in database", in.Subject)
		if !user.Enabled {
			msg := fmt.Sprintf("User '%d' is not enabled", user.ID)
			log.Printf("[DEBUG] %s", msg)
			return nil, false, model.NewError(model.ErrOther, msg)
		}
		// We got a user
		log.Printf("[DEBUG] Returning enabled user '%#v'", user)
		return &user, false, nil
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
	return &user, true, translateError(err, nil, nil, "")
}

func (p PostgresPersister) SelectUsersByID(ctx context.Context, ids []uint32) ([]model.User, error) {
	users := make([]model.User, 0)

	rows, err := p.db.QueryContext(ctx, "SELECT id, body, insert_time, last_update_time FROM cms_user "+
		"WHERE id = ANY($1)", pq.Array(ids))
	if err != nil {
		return nil, translateError(err, nil, nil, "")
	}
	defer rows.Close()
	for rows.Next() {
		var user model.User
		err := rows.Scan(
			&user.ID,
			&user.UserBody,
			&user.InsertTime,
			&user.LastUpdateTime,
		)
		if err != nil {
			return nil, translateError(err, nil, nil, "")
		}
		users = append(users, user)
	}
	return users, nil
}
