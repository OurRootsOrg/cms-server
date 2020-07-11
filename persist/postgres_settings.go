package persist

import (
	"context"
	"database/sql"

	"github.com/ourrootsorg/cms-server/model"
)

const SettingsID = 1

// SelectSettings selects the Settings object if it exists or returns ErrNoRows
func (p PostgresPersister) SelectSettings(ctx context.Context) (*model.Settings, *model.Error) {
	var settings model.Settings
	err := p.db.QueryRowContext(ctx, "SELECT body, insert_time, last_update_time FROM settings WHERE id=$1", SettingsID).Scan(
		&settings.SettingsBody,
		&settings.InsertTime,
		&settings.LastUpdateTime,
	)
	id := uint32(SettingsID)
	return &settings, translateError(err, &id, nil, "")
}

// UpsertSettings updates or inserts a Settings object in the database and returns the updated Settings
func (p PostgresPersister) UpsertSettings(ctx context.Context, in model.Settings) (*model.Settings, *model.Error) {
	var settings model.Settings
	err := p.db.QueryRowContext(ctx,
		`UPDATE settings SET body = $1, last_update_time = CURRENT_TIMESTAMP
		 WHERE id = $2 AND last_update_time = $3
		 RETURNING body, insert_time, last_update_time`,
		in.SettingsBody, SettingsID, in.LastUpdateTime).
		Scan(
			&settings.SettingsBody,
			&settings.InsertTime,
			&settings.LastUpdateTime,
		)
	if err != nil && err == sql.ErrNoRows {
		// Either non-existent or last_update_time didn't match
		s, e := p.SelectSettings(ctx)
		if e == nil {
			// Row exists, so it must be a non-matching update time
			return nil, model.NewError(model.ErrConcurrentUpdate, s.LastUpdateTime.String(), in.LastUpdateTime.String())
		} else if e.Code == model.ErrNotFound {
			// row doesn't exist; need to insert
			err := p.db.QueryRowContext(ctx,
				`INSERT INTO settings (id, body)
				VALUES ($1, $2)
		 		RETURNING body, insert_time, last_update_time`,
				SettingsID, in.SettingsBody).
				Scan(
					&settings.SettingsBody,
					&settings.InsertTime,
					&settings.LastUpdateTime,
				)
			id := uint32(SettingsID)
			return &settings, translateError(err, &id, nil, "")
		}
	}
	id := uint32(SettingsID)
	return &settings, translateError(err, &id, nil, "")
}
