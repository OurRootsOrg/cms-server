package persist

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/lib/pq"
	"github.com/ourrootsorg/cms-server/model"
)

// SelectRecordsForPost selects all records for a post
func (p PostgresPersister) SelectRecordsForPost(ctx context.Context, postID uint32) ([]model.Record, error) {
	rows, err := p.db.QueryContext(ctx, "SELECT id, post_id, body, ix_hash, insert_time, last_update_time FROM record WHERE post_id=$1", postID)
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()
	records := make([]model.Record, 0)
	for rows.Next() {
		var record model.Record
		err := rows.Scan(&record.ID, &record.Post, &record.RecordBody, &record.IxHash, &record.InsertTime, &record.LastUpdateTime)
		if err != nil {
			return nil, translateError(err)
		}
		records = append(records, record)
	}
	return records, nil
}

// SelectRecordsByID selects many records
func (p PostgresPersister) SelectRecordsByID(ctx context.Context, ids []uint32) ([]model.Record, error) {
	records := make([]model.Record, 0)
	if len(ids) == 0 {
		return records, nil
	}
	rows, err := p.db.QueryContext(ctx, "SELECT id, post_id, body, ix_hash, insert_time, last_update_time FROM record WHERE id = ANY($1)", pq.Array(ids))
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()
	for rows.Next() {
		var record model.Record
		err := rows.Scan(&record.ID, &record.Post, &record.RecordBody, &record.IxHash, &record.InsertTime, &record.LastUpdateTime)
		if err != nil {
			return nil, translateError(err)
		}
		records = append(records, record)
	}
	return records, nil
}

// SelectOneRecord selects a single record
func (p PostgresPersister) SelectOneRecord(ctx context.Context, id uint32) (*model.Record, error) {
	var record model.Record
	err := p.db.QueryRowContext(ctx, "SELECT id, post_id, body, ix_hash, insert_time, last_update_time FROM record WHERE id=$1", id).Scan(
		&record.ID,
		&record.Post,
		&record.RecordBody,
		&record.IxHash,
		&record.InsertTime,
		&record.LastUpdateTime,
	)
	if err != nil {
		return nil, translateError(err)
	}
	return &record, nil
}

// InsertRecord inserts a RecordBody into the database and returns the inserted Record
func (p PostgresPersister) InsertRecord(ctx context.Context, in model.RecordIn) (*model.Record, error) {
	var record model.Record
	err := p.db.QueryRowContext(ctx,
		`INSERT INTO record (post_id, body)
		 VALUES ($1, $2)
		 RETURNING id, post_id, body, ix_hash, insert_time, last_update_time`,
		in.Post, in.RecordBody).
		Scan(
			&record.ID,
			&record.Post,
			&record.RecordBody,
			&record.IxHash,
			&record.InsertTime,
			&record.LastUpdateTime,
		)
	return &record, translateError(err)
}

// UpdateRecord updates a Record in the database and returns the updated Record
func (p PostgresPersister) UpdateRecord(ctx context.Context, id uint32, in model.Record) (*model.Record, error) {
	var record model.Record
	err := p.db.QueryRowContext(ctx,
		`UPDATE record SET body = $1, post_id = $2, ix_hash = $3, last_update_time = CURRENT_TIMESTAMP
		 WHERE id = $4 AND last_update_time = $5
		 RETURNING id, post_id, body, ix_hash, insert_time, last_update_time`,
		in.RecordBody, in.Post, in.IxHash, id, in.LastUpdateTime).
		Scan(
			&record.ID,
			&record.Post,
			&record.RecordBody,
			&record.IxHash,
			&record.InsertTime,
			&record.LastUpdateTime,
		)
	if err != nil && err == sql.ErrNoRows {
		// Either non-existent or last_update_time didn't match
		c, _ := p.SelectOneRecord(ctx, id)
		if c != nil && c.ID == id {
			// Row exists, so it must be a non-matching update time
			return nil, model.NewError(model.ErrConcurrentUpdate, c.LastUpdateTime.String(), in.LastUpdateTime.String())
		}
		return nil, model.NewError(model.ErrNotFound, strconv.Itoa(int(id)))
	}
	return &record, translateError(err)
}

// DeleteRecord deletes a Record
func (p PostgresPersister) DeleteRecord(ctx context.Context, id uint32) error {
	_, err := p.db.ExecContext(ctx, "DELETE FROM record WHERE id = $1", id)
	return translateError(err)
}

// DeleteRecordsForPost deletes all Records for a post
func (p PostgresPersister) DeleteRecordsForPost(ctx context.Context, postID uint32) error {
	_, err := p.db.ExecContext(ctx, "DELETE FROM record WHERE post_id = $1", postID)
	return translateError(err)
}
