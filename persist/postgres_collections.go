package persist

import (
	"context"
	"database/sql"
	"fmt"

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
		var dbID int32
		var categoryID int32
		var collection model.Collection
		err := rows.Scan(&dbID, &categoryID, &collection.CollectionBody, &collection.InsertTime, &collection.LastUpdateTime)
		if err != nil {
			return nil, translateError(err)
		}
		collection.ID = model.MakeCollectionID(dbID)
		collection.Category = model.MakeCategoryID(categoryID)
		collections = append(collections, collection)
	}
	return collections, nil
}

// SelectOneCollection selects a single collection
func (p PostgresPersister) SelectOneCollection(ctx context.Context, id string) (model.Collection, error) {
	var collection model.Collection
	var dbID int32
	var catID int32
	n, err := fmt.Sscanf(id+"\n", p.pathPrefix+model.CollectionIDFormat+"\n", &dbID)
	if err != nil || n != 1 {
		// Bad ID
		return collection, model.NewError(model.ErrNotFound, id)
	}
	err = p.db.QueryRowContext(ctx, "SELECT id, category_id, body, insert_time, last_update_time FROM collection WHERE id=$1", dbID).Scan(
		&dbID,
		&catID,
		&collection.CollectionBody,
		&collection.InsertTime,
		&collection.LastUpdateTime,
	)
	if err != nil {
		return collection, translateError(err)
	}
	collection.ID = model.MakeCollectionID(dbID)
	collection.Category = model.MakeCategoryID(catID)
	return collection, nil
}

// InsertCollection inserts a CollectionBody into the database and returns the inserted Collection
func (p PostgresPersister) InsertCollection(ctx context.Context, in model.CollectionIn) (model.Collection, error) {
	var dbID int32
	var collection model.Collection
	var catID int32
	n, err := fmt.Sscanf(in.Category, p.pathPrefix+model.CategoryIDFormat, &catID)
	if err != nil || n != 1 {
		// Bad ID
		return collection, model.NewError(model.ErrBadReference, in.Category, "category")
	}
	err = p.db.QueryRowContext(ctx,
		`INSERT INTO collection (category_id, body) 
		 VALUES ($1, $2) 
		 RETURNING id, category_id, body, insert_time, last_update_time`,
		catID, in.CollectionBody).
		Scan(
			&dbID,
			&catID,
			&collection.CollectionBody,
			&collection.InsertTime,
			&collection.LastUpdateTime,
		)
	collection.ID = model.MakeCollectionID(dbID)
	collection.Category = model.MakeCategoryID(catID)
	return collection, translateError(err)
}

// UpdateCollection updates a Collection in the database and returns the updated Collection
func (p PostgresPersister) UpdateCollection(ctx context.Context, id string, in model.Collection) (model.Collection, error) {
	var collection model.Collection
	var dbID int32
	n, err := fmt.Sscanf(id+"\n", p.pathPrefix+model.CollectionIDFormat+"\n", &dbID)
	if err != nil || n != 1 {
		// Bad ID
		return collection, model.NewError(model.ErrNotFound, id)
	}
	var catID int32
	n, err = fmt.Sscanf(in.Category, p.pathPrefix+model.CategoryIDFormat, &catID)
	if err != nil || n != 1 {
		// Bad ID
		return collection, model.NewError(model.ErrBadReference, in.Category, "category")
	}
	err = p.db.QueryRowContext(ctx,
		`UPDATE collection SET body = $1, category_id = $2, last_update_time = CURRENT_TIMESTAMP 
		 WHERE id = $3 AND last_update_time = $4
		 RETURNING id, category_id, body, insert_time, last_update_time`,
		in.CollectionBody, catID, dbID, in.LastUpdateTime).
		Scan(
			&dbID,
			&catID,
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
		return collection, model.NewError(model.ErrNotFound, id)
	}
	collection.ID = model.MakeCollectionID(dbID)
	collection.Category = model.MakeCategoryID(catID)
	return collection, translateError(err)
}

// DeleteCollection deletes a Collection
func (p PostgresPersister) DeleteCollection(ctx context.Context, id string) error {
	var dbID int32
	n, err := fmt.Sscanf(id+"\n", p.pathPrefix+model.CollectionIDFormat+"\n", &dbID)
	if err != nil || n != 1 {
		// Bad ID
		return model.NewError(model.ErrNotFound, id)
	}
	_, err = p.db.ExecContext(ctx, "DELETE FROM collection WHERE id = $1", dbID)
	return translateError(err)
}
