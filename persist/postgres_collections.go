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
	rows, err := p.db.QueryContext(ctx,
		`SELECT id, array_agg(cc.category_id), body, insert_time, last_update_time
			   FROM collection LEFT JOIN collection_category cc ON id = cc.collection_id GROUP BY id`)
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()
	collections := make([]model.Collection, 0)
	for rows.Next() {
		var categories []int64
		var collection model.Collection
		err := rows.Scan(&collection.ID, pq.Array(&categories), &collection.CollectionBody, &collection.InsertTime, &collection.LastUpdateTime)
		if err != nil {
			return nil, translateError(err)
		}
		collection.Categories = make([]uint32, len(categories))
		for i, cat := range categories {
			collection.Categories[i] = uint32(cat)
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

	rows, err := p.db.QueryContext(ctx,
		`SELECT id, array_agg(cc.category_id), body, insert_time, last_update_time
			   FROM collection LEFT JOIN collection_category cc ON id = cc.collection_id WHERE id = ANY($1) GROUP BY id`, pq.Array(ids))
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()
	for rows.Next() {
		var categories []int64
		var collection model.Collection
		err := rows.Scan(&collection.ID, pq.Array(&categories), &collection.CollectionBody, &collection.InsertTime, &collection.LastUpdateTime)
		if err != nil {
			return nil, translateError(err)
		}
		collection.Categories = make([]uint32, len(categories))
		for i, cat := range categories {
			collection.Categories[i] = uint32(cat)
		}
		collections = append(collections, collection)
	}
	return collections, nil
}

// SelectOneCollection selects a single collection
func (p PostgresPersister) SelectOneCollection(ctx context.Context, id uint32) (model.Collection, error) {
	var categories []int64
	var collection model.Collection
	err := p.db.QueryRowContext(ctx,
		`SELECT id, array_agg(cc.category_id), body, insert_time, last_update_time
			   FROM collection LEFT JOIN collection_category cc ON id = cc.collection_id WHERE id = $1 GROUP BY id`, id).Scan(
		&collection.ID,
		pq.Array(&categories),
		&collection.CollectionBody,
		&collection.InsertTime,
		&collection.LastUpdateTime,
	)
	if err != nil {
		return collection, translateError(err)
	}
	collection.Categories = make([]uint32, len(categories))
	for i, cat := range categories {
		collection.Categories[i] = uint32(cat)
	}
	return collection, nil
}

// InsertCollection inserts a CollectionBody into the database and returns the inserted Collection
func (p PostgresPersister) InsertCollection(ctx context.Context, in model.CollectionIn) (model.Collection, error) {
	var collection model.Collection
	// create a transaction so collection and collection_category stay in sync
	tx, err := p.db.Begin()
	if err != nil {
		return collection, translateError(err)
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx,
		`INSERT INTO collection (body) 
		 VALUES ($1) 
		 RETURNING id, body, insert_time, last_update_time`,
		in.CollectionBody).
		Scan(
			&collection.ID,
			&collection.CollectionBody,
			&collection.InsertTime,
			&collection.LastUpdateTime,
		)
	if err != nil {
		return collection, translateError(err)
	}
	for _, category := range in.Categories {
		_, err = tx.ExecContext(ctx, "INSERT INTO collection_category (collection_id, category_id) VALUES ($1, $2)", collection.ID, category)
		if err != nil {
			return collection, translateError(err)
		}
	}
	collection.Categories = append([]uint32(nil), in.Categories...)
	err = tx.Commit()
	return collection, translateError(err)
}

// UpdateCollection updates a Collection in the database and returns the updated Collection
func (p PostgresPersister) UpdateCollection(ctx context.Context, id uint32, in model.Collection) (model.Collection, error) {
	var collection model.Collection
	// create a transaction so collection and collection_category stay in sync
	tx, err := p.db.Begin()
	if err != nil {
		return collection, translateError(err)
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx,
		`UPDATE collection SET body = $1, last_update_time = CURRENT_TIMESTAMP 
		 WHERE id = $2 AND last_update_time = $3
		 RETURNING id, body, insert_time, last_update_time`,
		in.CollectionBody, id, in.LastUpdateTime).
		Scan(
			&collection.ID,
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
	// delete and re-add categories (in the future we could calculate the differences and add/delete just what we need to)
	_, err = tx.ExecContext(ctx, "DELETE FROM collection_category WHERE collection_id = $1", id)
	if err != nil {
		return collection, translateError(err)
	}
	for _, category := range in.Categories {
		_, err = tx.ExecContext(ctx, "INSERT INTO collection_category (collection_id, category_id) VALUES ($1, $2)", collection.ID, category)
		if err != nil {
			return collection, translateError(err)
		}
	}
	collection.Categories = append([]uint32(nil), in.Categories...)
	err = tx.Commit()
	return collection, translateError(err)
}

// DeleteCollection deletes a Collection
func (p PostgresPersister) DeleteCollection(ctx context.Context, id uint32) error {
	// create a transaction so collection and collection_category stay in sync
	tx, err := p.db.Begin()
	if err != nil {
		return translateError(err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "DELETE FROM collection_category WHERE collection_id = $1", id)
	if err != nil {
		return translateError(err)
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM collection WHERE id = $1", id)
	if err != nil {
		return translateError(err)
	}
	err = tx.Commit()
	return translateError(err)
}
