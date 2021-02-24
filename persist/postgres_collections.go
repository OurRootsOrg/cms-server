package persist

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/ourrootsorg/cms-server/utils"

	"github.com/lib/pq"

	"github.com/ourrootsorg/cms-server/model"
)

// Collection persistence methods

// SelectCollections selects all collections
func (p PostgresPersister) SelectCollections(ctx context.Context) ([]model.Collection, error) {
	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := p.db.QueryContext(ctx,
		`SELECT id, array_agg(cc.category_id), body, insert_time, last_update_time
			   FROM collection LEFT JOIN collection_category cc ON id = cc.collection_id 
			   WHERE society_id=$1 GROUP BY id`, societyID)
	if err != nil {
		return nil, translateError(err, nil, nil, "")
	}
	defer rows.Close()
	collections := make([]model.Collection, 0)
	for rows.Next() {
		var categories []int64
		var collection model.Collection
		err := rows.Scan(&collection.ID, pq.Array(&categories), &collection.CollectionBody, &collection.InsertTime, &collection.LastUpdateTime)
		if err != nil {
			return nil, translateError(err, nil, nil, "")
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
	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	collections := make([]model.Collection, 0)
	if len(ids) == 0 {
		return collections, nil
	}

	rows, err := p.db.QueryContext(ctx,
		`SELECT id, array_agg(cc.category_id), body, insert_time, last_update_time
			   FROM collection LEFT JOIN collection_category cc ON id = cc.collection_id 
			   WHERE society_id=$1 AND id = ANY($2) GROUP BY id`, societyID, pq.Array(ids))
	if err != nil {
		return nil, translateError(err, nil, nil, "")
	}
	defer rows.Close()
	for rows.Next() {
		var categories []int64
		var collection model.Collection
		err := rows.Scan(&collection.ID, pq.Array(&categories), &collection.CollectionBody, &collection.InsertTime, &collection.LastUpdateTime)
		if err != nil {
			return nil, translateError(err, nil, nil, "")
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
func (p PostgresPersister) SelectOneCollection(ctx context.Context, id uint32) (*model.Collection, error) {
	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var categories []int64
	var collection model.Collection
	err = p.db.QueryRowContext(ctx,
		`SELECT id, array_agg(cc.category_id), body, insert_time, last_update_time
			   FROM collection LEFT JOIN collection_category cc ON id = cc.collection_id 
			   WHERE society_id=$1 AND id = $2 GROUP BY id`, societyID, id).Scan(
		&collection.ID,
		pq.Array(&categories),
		&collection.CollectionBody,
		&collection.InsertTime,
		&collection.LastUpdateTime,
	)
	if err != nil {
		return nil, translateError(err, &id, nil, "")
	}
	collection.Categories = make([]uint32, len(categories))
	for i, cat := range categories {
		collection.Categories[i] = uint32(cat)
	}
	return &collection, nil
}

// InsertCollection inserts a CollectionBody into the database and returns the inserted Collection
func (p PostgresPersister) InsertCollection(ctx context.Context, in model.CollectionIn) (*model.Collection, error) {
	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var collection model.Collection
	// create a transaction so collection and collection_category stay in sync
	tx, err := p.db.Begin()
	if err != nil {
		return nil, translateError(err, nil, nil, "")
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx,
		`INSERT INTO collection (society_id, body)
		 VALUES ($1, $2)
		 RETURNING id, body, insert_time, last_update_time`,
		societyID, in.CollectionBody).
		Scan(
			&collection.ID,
			&collection.CollectionBody,
			&collection.InsertTime,
			&collection.LastUpdateTime,
		)
	if err != nil {
		return nil, translateError(err, nil, nil, "")
	}
	for _, category := range in.Categories {
		_, err = tx.ExecContext(ctx, "INSERT INTO collection_category (collection_id, category_id) VALUES ($1, $2)", collection.ID, category)
		if err != nil {
			return nil, translateError(err, &collection.ID, &category, "category")
		}
	}
	collection.Categories = append([]uint32(nil), in.Categories...)
	err = tx.Commit()
	return &collection, translateError(err, nil, nil, "")
}

// UpdateCollection updates a Collection in the database and returns the updated Collection
func (p PostgresPersister) UpdateCollection(ctx context.Context, id uint32, in model.Collection) (*model.Collection, error) {
	var collection model.Collection
	// create a transaction so collection and collection_category stay in sync
	tx, err := p.db.Begin()
	if err != nil {
		return nil, translateError(err, &id, nil, "")
	}
	defer tx.Rollback()

	// make sure collection exists for this society
	existingCollection, err := p.SelectOneCollection(ctx, id)
	if err != nil || existingCollection == nil || existingCollection.ID != id {
		return nil, model.NewError(model.ErrNotFound, strconv.Itoa(int(id)))
	}

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
		// Row exists, so it must be a non-matching update time
		return nil, model.NewError(model.ErrConcurrentUpdate, existingCollection.LastUpdateTime.String(), in.LastUpdateTime.String())
	}
	// delete and re-add categories (in the future we could calculate the differences and add/delete just what we need to)
	_, err = tx.ExecContext(ctx, "DELETE FROM collection_category WHERE collection_id = $1", id)
	if err != nil {
		return nil, translateError(err, &id, nil, "")
	}
	for _, category := range in.Categories {
		_, err = tx.ExecContext(ctx, "INSERT INTO collection_category (collection_id, category_id) VALUES ($1, $2)", collection.ID, category)
		if err != nil {
			return nil, translateError(err, &collection.ID, &category, "category")
		}
	}
	collection.Categories = append([]uint32(nil), in.Categories...)
	err = tx.Commit()
	return &collection, translateError(err, nil, nil, "")
}

// DeleteCollection deletes a Collection
func (p PostgresPersister) DeleteCollection(ctx context.Context, id uint32) error {
	// create a transaction so collection and collection_category stay in sync
	tx, err := p.db.Begin()
	if err != nil {
		return translateError(err, &id, nil, "")
	}
	defer tx.Rollback()

	// make sure collection exists for this society
	existingCollection, err := p.SelectOneCollection(ctx, id)
	if err != nil || existingCollection == nil || existingCollection.ID != id {
		return translateError(err, &id, nil, "")
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM collection_category WHERE collection_id = $1", id)
	if err != nil {
		return translateError(err, &id, nil, "")
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM collection WHERE id = $1", id)
	if err != nil {
		return translateError(err, &id, nil, "")
	}
	err = tx.Commit()
	return translateError(err, &id, nil, "")
}
