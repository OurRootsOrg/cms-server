package persist

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ourrootsorg/cms-server/model"
)

// SelectRecords selects all records
func (p PostgresPersister) SelectRecordsForPost(ctx context.Context, postID string) ([]model.Record, error) {
	var dbID int32
	n, err := fmt.Sscanf(postID+"\n", p.pathPrefix+model.PostIDFormat+"\n", &dbID)
	if err != nil || n != 1 {
		// Bad ID
		return nil, model.NewError(model.ErrBadReference, postID, "post")
	}
	rows, err := p.db.QueryContext(ctx, "SELECT id, post_id, body, ix_hash, insert_time, last_update_time FROM record WHERE post_id=$1", dbID)
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()
	records := make([]model.Record, 0)
	for rows.Next() {
		var dbID int32
		var postID int32
		var record model.Record
		err := rows.Scan(&dbID, &postID, &record.RecordBody, &record.IxHash, &record.InsertTime, &record.LastUpdateTime)
		if err != nil {
			return nil, translateError(err)
		}
		record.ID = model.MakeRecordID(dbID)
		record.Post = model.MakePostID(postID)
		records = append(records, record)
	}
	return records, nil
}

// SelectOneRecord selects a single record
func (p PostgresPersister) SelectOneRecord(ctx context.Context, id string) (model.Record, error) {
	var record model.Record
	var dbID int32
	var postID int32
	n, err := fmt.Sscanf(id+"\n", p.pathPrefix+model.RecordIDFormat+"\n", &dbID)
	if err != nil || n != 1 {
		// Bad ID
		return record, model.NewError(model.ErrNotFound, id)
	}
	err = p.db.QueryRowContext(ctx, "SELECT id, post_id, body, ix_hash, insert_time, last_update_time FROM record WHERE id=$1", dbID).Scan(
		&dbID,
		&postID,
		&record.RecordBody,
		&record.IxHash,
		&record.InsertTime,
		&record.LastUpdateTime,
	)
	if err != nil {
		return record, translateError(err)
	}
	record.ID = model.MakeRecordID(dbID)
	record.Post = model.MakePostID(postID)
	return record, nil
}

// InsertRecord inserts a RecordBody into the database and returns the inserted Record
func (p PostgresPersister) InsertRecord(ctx context.Context, in model.RecordIn) (model.Record, error) {
	var dbID int32
	var record model.Record
	var postID int32
	n, err := fmt.Sscanf(in.Post+"\n", p.pathPrefix+model.PostIDFormat+"\n", &postID)
	if err != nil || n != 1 {
		// Bad ID
		return record, model.NewError(model.ErrBadReference, in.Post, "post")
	}
	err = p.db.QueryRowContext(ctx,
		`INSERT INTO record (post_id, body) 
		 VALUES ($1, $2) 
		 RETURNING id, post_id, body, ix_hash, insert_time, last_update_time`,
		postID, in.RecordBody).
		Scan(
			&dbID,
			&postID,
			&record.RecordBody,
			&record.IxHash,
			&record.InsertTime,
			&record.LastUpdateTime,
		)
	record.ID = model.MakeRecordID(dbID)
	record.Post = model.MakePostID(postID)
	return record, translateError(err)
}

// UpdateRecord updates a Record in the database and returns the updated Record
func (p PostgresPersister) UpdateRecord(ctx context.Context, id string, in model.Record) (model.Record, error) {
	var record model.Record
	var dbID int32
	n, err := fmt.Sscanf(id+"\n", p.pathPrefix+model.RecordIDFormat+"\n", &dbID)
	if err != nil || n != 1 {
		// Bad ID
		return record, model.NewError(model.ErrNotFound, id)
	}
	var postID int32
	n, err = fmt.Sscanf(in.Post+"\n", p.pathPrefix+model.PostIDFormat+"\n", &postID)
	if err != nil || n != 1 {
		// Bad ID
		return record, model.NewError(model.ErrBadReference, in.Post, "post")
	}
	err = p.db.QueryRowContext(ctx,
		`UPDATE record SET body = $1, post_id = $2, ix_hash = $3, last_update_time = CURRENT_TIMESTAMP 
		 WHERE id = $4 AND last_update_time = $5
		 RETURNING id, post_id, body, ix_hash, insert_time, last_update_time`,
		in.RecordBody, postID, in.IxHash, dbID, in.LastUpdateTime).
		Scan(
			&dbID,
			&postID,
			&record.RecordBody,
			&record.IxHash,
			&record.InsertTime,
			&record.LastUpdateTime,
		)
	if err != nil && err == sql.ErrNoRows {
		// Either non-existent or last_update_time didn't match
		c, _ := p.SelectOneRecord(ctx, id)
		if c.ID == id {
			// Row exists, so it must be a non-matching update time
			return record, model.NewError(model.ErrConcurrentUpdate, c.LastUpdateTime.String(), in.LastUpdateTime.String())
		}
		return record, model.NewError(model.ErrNotFound, id)
	}
	record.ID = model.MakeRecordID(dbID)
	record.Post = model.MakePostID(postID)
	return record, translateError(err)
}

// DeleteRecord deletes a Record
func (p PostgresPersister) DeleteRecord(ctx context.Context, id string) error {
	var dbID int32
	n, err := fmt.Sscanf(id+"\n", p.pathPrefix+model.RecordIDFormat+"\n", &dbID)
	if err != nil || n != 1 {
		// Bad ID
		return model.NewError(model.ErrNotFound, id)
	}
	_, err = p.db.ExecContext(ctx, "DELETE FROM record WHERE id = $1", dbID)
	return translateError(err)
}

// DeleteRecordsForPost deletes all Records for a post
func (p PostgresPersister) DeleteRecordsForPost(ctx context.Context, postID string) error {
	var dbID int32
	n, err := fmt.Sscanf(postID+"\n", p.pathPrefix+model.PostIDFormat+"\n", &dbID)
	if err != nil || n != 1 {
		// Bad ID
		return model.NewError(model.ErrBadReference, postID, "post")
	}
	_, err = p.db.ExecContext(ctx, "DELETE FROM record WHERE post_id = $1", dbID)
	return translateError(err)
}
