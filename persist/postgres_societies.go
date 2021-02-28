package persist

import (
	"context"
	"database/sql"
	"log"
	"strconv"

	"github.com/lib/pq"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/utils"
)

// SelectSocietySummariesByID selects multiple society summaries
func (p PostgresPersister) SelectSocietySummariesByID(ctx context.Context, ids []uint32) ([]model.SocietySummary, error) {
	societySummaries := make([]model.SocietySummary, 0)

	rows, err := p.db.QueryContext(ctx, "SELECT id, body FROM society "+
		"WHERE id = ANY($1)", pq.Array(ids))
	if err != nil {
		return nil, translateError(err, nil, nil, "")
	}
	defer rows.Close()
	for rows.Next() {
		var society model.Society
		err := rows.Scan(&society.ID, &society.SocietyBody)
		if err != nil {
			return nil, translateError(err, nil, nil, "")
		}
		societySummary := createSocietySummary(&society)
		societySummaries = append(societySummaries, *societySummary)
	}
	return societySummaries, nil
}

// SelectOneSociety loads a SocietySummary from the database
func (p PostgresPersister) SelectSocietySummary(ctx context.Context, id uint32) (*model.SocietySummary, error) {
	var society model.Society
	log.Printf("[DEBUG] id: %d", id)
	err := p.db.QueryRowContext(ctx, "SELECT id, body FROM society "+
		"WHERE id=$1", id).Scan(
		&society.ID,
		&society.SocietyBody,
	)
	if err != nil {
		return nil, translateError(err, &id, nil, "")
	}
	return createSocietySummary(&society), nil
}

// SelectOneSociety loads the current Society from the database
func (p PostgresPersister) SelectSociety(ctx context.Context, id uint32) (*model.Society, error) {
	var society model.Society
	log.Printf("[DEBUG] id: %d", id)
	err := p.db.QueryRowContext(ctx, "SELECT id, body, insert_time, last_update_time FROM society "+
		"WHERE id=$1", id).Scan(
		&society.ID,
		&society.SocietyBody,
		&society.InsertTime,
		&society.LastUpdateTime,
	)
	if err != nil {
		return nil, translateError(err, &id, nil, "")
	}
	return &society, nil
}

// InsertSociety inserts a SocietyBody into the database and returns the inserted Society
func (p PostgresPersister) InsertSociety(ctx context.Context, in model.SocietyIn) (*model.Society, error) {
	var society model.Society
	row := p.db.QueryRowContext(ctx, "INSERT INTO society (body) VALUES ($1) "+
		"RETURNING id, body, insert_time, last_update_time", in)
	err := row.Scan(
		&society.ID,
		&society.SocietyBody,
		&society.InsertTime,
		&society.LastUpdateTime,
	)
	return &society, translateError(err, nil, nil, "")
}

// UpdateSociety updates a Society in the database and returns the updated Society
func (p PostgresPersister) UpdateSociety(ctx context.Context, in model.Society) (*model.Society, error) {
	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var society model.Society
	err = p.db.QueryRowContext(ctx, "UPDATE society SET body = $1, last_update_time = CURRENT_TIMESTAMP "+
		"WHERE id = $2 AND last_update_time = $3 RETURNING id, body, insert_time, last_update_time",
		in.SocietyBody, societyID, in.LastUpdateTime).
		Scan(
			&society.ID,
			&society.SocietyBody,
			&society.InsertTime,
			&society.LastUpdateTime,
		)
	if err != nil && err == sql.ErrNoRows {
		// Either non-existent or last_update_time didn't match
		c, _ := p.SelectSociety(ctx, societyID)
		if c != nil && c.ID == societyID {
			// Row exists, so it must be a non-matching update time
			return nil, model.NewError(model.ErrConcurrentUpdate, c.LastUpdateTime.String(), in.LastUpdateTime.String())
		}
		return nil, model.NewError(model.ErrNotFound, strconv.Itoa(int(societyID)))
	}
	return &society, translateError(err, &societyID, nil, "")
}

// DeleteSociety deletes a society
func (p PostgresPersister) DeleteSociety(ctx context.Context) error {
	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		return err
	}
	_, err = p.db.ExecContext(ctx, "DELETE FROM society WHERE id = $1", societyID)
	return translateError(err, &societyID, nil, "")
}

func createSocietySummary(society *model.Society) *model.SocietySummary {
	return &model.SocietySummary{
		ID:           society.ID,
		Name:         society.Name,
		PostMetadata: society.PostMetadata,
	}
}
