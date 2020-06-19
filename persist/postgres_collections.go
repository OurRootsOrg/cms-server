package persist

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/lib/pq"

	"github.com/ourrootsorg/cms-server/model"
)

// Collection persistence methods

// SelectCollections selects all collections
func (p PostgresPersister) SelectCollections(ctx context.Context) ([]model.Collection, error) {
	rows, err := p.db.QueryContext(ctx, "SELECT id, category_id, body, insert_time, last_update_time FROM collection")
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()
	collections := make([]model.Collection, 0)
	for rows.Next() {
		var collection model.Collection
		err := rows.Scan(&collection.ID, &collection.Category, &collection.CollectionBody, &collection.InsertTime, &collection.LastUpdateTime)
		if err != nil {
			return nil, translateError(err)
		}
		collections = append(collections, collection)
	}
	return collections, nil
}

// SelectCollectionsByID selects many collections
func (p PostgresPersister) SelectCollectionsByID(ctx context.Context, ids []uint32) ([]model.Collection, error) {
	collections := make([]model.Collection, 0)
	if len(ids) == 0 {
		return collections, nil
	}

	rows, err := p.db.QueryContext(ctx, "SELECT id, category_id, body, insert_time, last_update_time FROM collection WHERE id = ANY($1)", pq.Array(ids))
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()
	for rows.Next() {
		var collection model.Collection
		err := rows.Scan(&collection.ID, &collection.Category, &collection.CollectionBody, &collection.InsertTime, &collection.LastUpdateTime)
		if err != nil {
			return nil, translateError(err)
		}
		collections = append(collections, collection)
	}
	return collections, nil
}

// SelectOneCollection selects a single collection
func (p PostgresPersister) SelectOneCollection(ctx context.Context, id uint32) (model.Collection, error) {
	var collection model.Collection
	err := p.db.QueryRowContext(ctx, "SELECT id, category_id, body, insert_time, last_update_time FROM collection WHERE id=$1", id).Scan(
		&collection.ID,
		&collection.Category,
		&collection.CollectionBody,
		&collection.InsertTime,
		&collection.LastUpdateTime,
	)
	if err != nil {
		return collection, translateError(err)
	}
	return collection, nil
}

// InsertCollection inserts a CollectionBody into the database and returns the inserted Collection
func (p PostgresPersister) InsertCollection(ctx context.Context, in model.CollectionIn) (model.Collection, error) {
	var collection model.Collection
	err := p.db.QueryRowContext(ctx,
		`INSERT INTO collection (category_id, body) 
		 VALUES ($1, $2) 
		 RETURNING id, category_id, body, insert_time, last_update_time`,
		in.Category, in.CollectionBody).
		Scan(
			&collection.ID,
			&collection.Category,
			&collection.CollectionBody,
			&collection.InsertTime,
			&collection.LastUpdateTime,
		)
	return collection, translateError(err)
}

// UpdateCollection updates a Collection in the database and returns the updated Collection
func (p PostgresPersister) UpdateCollection(ctx context.Context, id uint32, in model.Collection) (model.Collection, error) {
	var collection model.Collection
	err := p.db.QueryRowContext(ctx,
		`UPDATE collection SET body = $1, category_id = $2, last_update_time = CURRENT_TIMESTAMP 
		 WHERE id = $3 AND last_update_time = $4
		 RETURNING id, category_id, body, insert_time, last_update_time`,
		in.CollectionBody, in.Category, id, in.LastUpdateTime).
		Scan(
			&collection.ID,
			&collection.Category,
			&collection.CollectionBody,
			&collection.InsertTime,
			&collection.LastUpdateTime,
		)
	if err != nil && err == sql.ErrNoRows {
		// Either non-existent or last_update_time didn't match
		c, _ := p.SelectOneCollection(ctx, id)
		if c.ID == id {
			// Row exists, so it must be a non-matching update time
			return collection, model.NewError(model.ErrConcurrentUpdate, c.LastUpdateTime.String(), in.LastUpdateTime.String())
		}
		return collection, model.NewError(model.ErrNotFound, strconv.Itoa(int(id)))
	}
	return collection, translateError(err)
}

// DeleteCollection deletes a Collection
func (p PostgresPersister) DeleteCollection(ctx context.Context, id uint32) error {
	_, err := p.db.ExecContext(ctx, "DELETE FROM collection WHERE id = $1", id)
	return translateError(err)
}
