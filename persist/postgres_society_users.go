package persist

import (
	"context"
	"database/sql"
	"log"
	"strconv"

	"github.com/ourrootsorg/cms-server/utils"

	"github.com/ourrootsorg/cms-server/model"
)

// SelectSocietyUsers selects all SocietyUsers for current society
func (p PostgresPersister) SelectSocietyUsers(ctx context.Context) ([]model.SocietyUser, error) {
	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	societyUsers := make([]model.SocietyUser, 0)

	rows, err := p.db.QueryContext(ctx,
		"SELECT id, body, user_id, society_id, insert_time, last_update_time FROM society_user "+
			"WHERE society_id = $1", societyID)
	if err != nil {
		return nil, translateError(err, nil, nil, "")
	}
	defer rows.Close()
	for rows.Next() {
		var societyUser model.SocietyUser
		err := rows.Scan(&societyUser.ID, &societyUser.SocietyUserBody, &societyUser.UserID, &societyUser.SocietyID, &societyUser.InsertTime, &societyUser.LastUpdateTime)
		if err != nil {
			return nil, translateError(err, nil, nil, "")
		}
		societyUsers = append(societyUsers, societyUser)
	}
	return societyUsers, nil
}

func (p PostgresPersister) SelectAllSocietyUsersByUser(ctx context.Context, userID uint32) ([]model.SocietyUser, error) {
	societyUsers := make([]model.SocietyUser, 0)

	rows, err := p.db.QueryContext(ctx,
		"SELECT id, body, user_id, society_id, insert_time, last_update_time FROM society_user "+
			"WHERE user_id = $1", userID)
	if err != nil {
		return nil, translateError(err, nil, nil, "")
	}
	defer rows.Close()
	for rows.Next() {
		var societyUser model.SocietyUser
		err := rows.Scan(&societyUser.ID, &societyUser.SocietyUserBody, &societyUser.UserID, &societyUser.SocietyID, &societyUser.InsertTime, &societyUser.LastUpdateTime)
		if err != nil {
			return nil, translateError(err, nil, nil, "")
		}
		societyUsers = append(societyUsers, societyUser)
	}
	return societyUsers, nil
}

// SelectOneSocietyUser loads a single SocietyUser from the database
func (p PostgresPersister) SelectOneSocietyUser(ctx context.Context, id uint32) (*model.SocietyUser, error) {
	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var societyUser model.SocietyUser
	log.Printf("[DEBUG] id: %d", id)
	err = p.db.QueryRowContext(ctx, "SELECT id, body, user_id, society_id, insert_time, last_update_time FROM society_user "+
		"WHERE society_id=$1 AND id=$2", societyID, id).Scan(
		&societyUser.ID,
		&societyUser.SocietyUserBody,
		&societyUser.UserID,
		&societyUser.SocietyID,
		&societyUser.InsertTime,
		&societyUser.LastUpdateTime,
	)
	if err != nil {
		return nil, translateError(err, &id, nil, "")
	}
	return &societyUser, nil
}

func (p PostgresPersister) SelectOneSocietyUserByUser(ctx context.Context, userID uint32) (*model.SocietyUser, error) {
	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var societyUser model.SocietyUser
	log.Printf("[DEBUG] userID: %d", userID)
	err = p.db.QueryRowContext(ctx, "SELECT id, body, user_id, society_id, insert_time, last_update_time FROM society_user "+
		"WHERE society_id=$1 AND user_id=$2", societyID, userID).Scan(
		&societyUser.ID,
		&societyUser.SocietyUserBody,
		&societyUser.UserID,
		&societyUser.SocietyID,
		&societyUser.InsertTime,
		&societyUser.LastUpdateTime,
	)
	if err != nil {
		return nil, translateError(err, &userID, nil, "")
	}
	return &societyUser, nil
}

// InsertSocietyUser inserts a SocietyUserBody into the database and returns the inserted SocietyUser
func (p PostgresPersister) InsertSocietyUser(ctx context.Context, in model.SocietyUserIn) (*model.SocietyUser, error) {
	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var societyUser model.SocietyUser
	row := p.db.QueryRowContext(ctx, "INSERT INTO society_user (body, user_id, society_id) VALUES ($1, $2, $3) "+
		"RETURNING id, body, user_id, society_id, insert_time, last_update_time", in.SocietyUserBody, in.UserID, societyID)
	err = row.Scan(
		&societyUser.ID,
		&societyUser.SocietyUserBody,
		&societyUser.UserID,
		&societyUser.SocietyID,
		&societyUser.InsertTime,
		&societyUser.LastUpdateTime,
	)
	return &societyUser, translateError(err, nil, nil, "")
}

// UpdateSocietyUser updates a SocietyUser in the database and returns the updated SocietyUser
// can't update societyID or userID
func (p PostgresPersister) UpdateSocietyUser(ctx context.Context, id uint32, in model.SocietyUser) (*model.SocietyUser, error) {
	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var societyUser model.SocietyUser
	err = p.db.QueryRowContext(ctx, "UPDATE society_user SET body = $1, last_update_time = CURRENT_TIMESTAMP "+
		"WHERE society_id = $3 AND id = $4 AND last_update_time = $5 RETURNING id, body, user_id, society_id, insert_time, last_update_time",
		in.SocietyUserBody, in.UserID, societyID, id, in.LastUpdateTime).
		Scan(
			&societyUser.ID,
			&societyUser.SocietyUserBody,
			&societyUser.UserID,
			&societyUser.SocietyID,
			&societyUser.InsertTime,
			&societyUser.LastUpdateTime,
		)
	if err != nil && err == sql.ErrNoRows {
		// Either non-existent or last_update_time didn't match
		c, _ := p.SelectOneSocietyUser(ctx, id)
		if c != nil && c.ID == id {
			// Row exists, so it must be a non-matching update time
			return nil, model.NewError(model.ErrConcurrentUpdate, c.LastUpdateTime.String(), in.LastUpdateTime.String())
		}
		return nil, model.NewError(model.ErrNotFound, strconv.Itoa(int(id)))
	}
	return &societyUser, translateError(err, &id, nil, "")
}

// DeleteSocietyUser deletes a SocietyUser
func (p PostgresPersister) DeleteSocietyUser(ctx context.Context, id uint32) error {
	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		return err
	}
	_, err = p.db.ExecContext(ctx, "DELETE FROM society_user WHERE society_id = $1 AND id = $2", societyID, id)
	return translateError(err, &id, nil, "")
}
