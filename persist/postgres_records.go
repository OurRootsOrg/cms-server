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
		return nil, translateError(err, &postID, nil, "")
	}
	defer rows.Close()
	records := make([]model.Record, 0)
	for rows.Next() {
		var record model.Record
		err := rows.Scan(&record.ID, &record.Post, &record.RecordBody, &record.IxHash, &record.InsertTime, &record.LastUpdateTime)
		if err != nil {
			return nil, translateError(err, &postID, nil, "")
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
		return nil, translateError(err, nil, nil, "")
	}
	defer rows.Close()
	for rows.Next() {
		var record model.Record
		err := rows.Scan(&record.ID, &record.Post, &record.RecordBody, &record.IxHash, &record.InsertTime, &record.LastUpdateTime)
		if err != nil {
			return nil, translateError(err, nil, nil, "")
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
		return nil, translateError(err, &id, nil, "")
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
	return &record, translateError(err, nil, &record.Post, "post")
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
	return &record, translateError(err, &id, &record.Post, "post")
}

// DeleteRecord deletes a Record
func (p PostgresPersister) DeleteRecord(ctx context.Context, id uint32) error {
	_, err := p.db.ExecContext(ctx, "DELETE FROM record WHERE id = $1", id)
	return translateError(err, &id, nil, "")
}

// DeleteRecordsForPost deletes all Records for a post
func (p PostgresPersister) DeleteRecordsForPost(ctx context.Context, postID uint32) error {
	_, err := p.db.ExecContext(ctx, "DELETE FROM record WHERE post_id = $1", postID)
	return translateError(err, &postID, nil, "")
}

// SelectRecordsForPost selects all records households for a post
func (p PostgresPersister) SelectRecordHouseholdsForPost(ctx context.Context, postID uint32) ([]model.RecordHousehold, error) {
	query := "SELECT post_id, household_id, record_ids, insert_time, last_update_time FROM record_household WHERE post_id = $1"
	rows, err := p.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, translateError(err, &postID, nil, "")
	}
	defer rows.Close()
	recordHouseholds := make([]model.RecordHousehold, 0)
	for rows.Next() {
		var recordHousehold model.RecordHousehold
		err := rows.Scan(&recordHousehold.Post, &recordHousehold.Household, &recordHousehold.Records, &recordHousehold.InsertTime, &recordHousehold.LastUpdateTime)
		if err != nil {
			return nil, translateError(err, &postID, nil, "")
		}
		recordHouseholds = append(recordHouseholds, recordHousehold)
	}
	return recordHouseholds, nil
}

// SelectOneRecordHousehold selects a single record household
func (p PostgresPersister) SelectOneRecordHousehold(ctx context.Context, postID uint32, householdID string) (*model.RecordHousehold, error) {
	var recordHousehold model.RecordHousehold
	query := "SELECT post_id, household_id, record_ids, insert_time, last_update_time FROM record_household WHERE post_id = $1 AND household_id = $2"
	err := p.db.QueryRowContext(ctx, query, postID, householdID).Scan(
		&recordHousehold.Post,
		&recordHousehold.Household,
		&recordHousehold.Records,
		&recordHousehold.InsertTime,
		&recordHousehold.LastUpdateTime,
	)
	if err != nil {
		return nil, translateError(err, &postID, nil, "")
	}
	return &recordHousehold, nil
}

// InsertRecordHousehold inserts a RecordHouseholdIn into the database and returns the inserted RecordHousehold
func (p PostgresPersister) InsertRecordHousehold(ctx context.Context, in model.RecordHouseholdIn) (*model.RecordHousehold, error) {
	var recordHousehold model.RecordHousehold
	err := p.db.QueryRowContext(ctx,
		`INSERT INTO record_household (post_id, household_id, record_ids)
		 VALUES ($1, $2, $3)
		 RETURNING post_id, household_id, record_ids, insert_time, last_update_time`,
		in.Post, in.Household, in.Records).
		Scan(
			&recordHousehold.Post,
			&recordHousehold.Household,
			&recordHousehold.Records,
			&recordHousehold.InsertTime,
			&recordHousehold.LastUpdateTime,
		)
	return &recordHousehold, translateError(err, nil, &in.Post, "post")
}

// DeleteRecordHouseholdsForPost deletes all Record Households for a post
func (p PostgresPersister) DeleteRecordHouseholdsForPost(ctx context.Context, postID uint32) error {
	_, err := p.db.ExecContext(ctx, "DELETE FROM record_household WHERE post_id = $1", postID)
	return translateError(err, &postID, nil, "")
}
