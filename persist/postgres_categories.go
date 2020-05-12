package persist

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/ourrootsorg/cms-server/model"
)

// SelectCategories loads all the categories from the database
func (p PostgresPersister) SelectCategories(ctx context.Context) ([]model.Category, error) {
	rows, err := p.db.QueryContext(ctx, "SELECT id, body, insert_time, last_update_time FROM category")
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()
	cats := make([]model.Category, 0)
	for rows.Next() {
		var dbID int32
		var cat model.Category
		err := rows.Scan(&dbID, &cat.CategoryBody, &cat.InsertTime, &cat.LastUpdateTime)
		if err != nil {
			return nil, translateError(err)
		}
		cat.ID = model.MakeCategoryID(dbID)
		cats = append(cats, cat)
	}
	return cats, nil
}

// SelectOneCategory loads a single category from the database
func (p PostgresPersister) SelectOneCategory(ctx context.Context, id string) (model.Category, error) {
	var cat model.Category
	var dbID int32
	n, err := fmt.Sscanf(id+"\n", p.pathPrefix+model.CategoryIDFormat+"\n", &dbID)
	if err != nil || n != 1 {
		// Bad ID
		return cat, model.NewError(model.ErrNotFound, id)
	}
	log.Printf("[DEBUG] id: %s, dbID: %d", id, dbID)
	err = p.db.QueryRowContext(ctx, "SELECT id, body, insert_time, last_update_time FROM category WHERE id=$1", dbID).Scan(
		&dbID,
		&cat.CategoryBody,
		&cat.InsertTime,
		&cat.LastUpdateTime,
	)
	if err != nil {
		return cat, translateError(err)
	}
	cat.ID = model.MakeCategoryID(dbID)
	return cat, nil
}

// InsertCategory inserts a CategoryBody into the database and returns the inserted Category
func (p PostgresPersister) InsertCategory(ctx context.Context, in model.CategoryIn) (model.Category, error) {
	var dbID int32
	var cat model.Category
	row := p.db.QueryRowContext(ctx, "INSERT INTO category (body) VALUES ($1) RETURNING id, body, insert_time, last_update_time", in)
	err := row.Scan(
		&dbID,
		&cat.CategoryBody,
		&cat.InsertTime,
		&cat.LastUpdateTime,
	)
	cat.ID = model.MakeCategoryID(dbID)
	return cat, translateError(err)
}

// UpdateCategory updates a Category in the database and returns the updated Category
func (p PostgresPersister) UpdateCategory(ctx context.Context, id string, in model.Category) (model.Category, error) {
	var cat model.Category
	var dbID int32
	n, err := fmt.Sscanf(id+"\n", p.pathPrefix+model.CategoryIDFormat+"\n", &dbID)
	if err != nil || n != 1 {
		// Bad ID
		return cat, model.NewError(model.ErrNotFound, id)
	}
	err = p.db.QueryRowContext(ctx, "UPDATE category SET body = $1, last_update_time = CURRENT_TIMESTAMP WHERE id = $2 AND last_update_time = $3 RETURNING id, body, insert_time, last_update_time", in.CategoryBody, dbID, in.LastUpdateTime).
		Scan(
			&dbID,
			&cat.CategoryBody,
			&cat.InsertTime,
			&cat.LastUpdateTime,
		)
	if err != nil && err == sql.ErrNoRows {
		// Either non-existent or last_update_time didn't match
		c, _ := p.SelectOneCategory(ctx, id)
		if c.ID == id {
			// Row exists, so it must be a non-matching update time
			return cat, model.NewError(model.ErrConcurrentUpdate, c.LastUpdateTime.String(), in.LastUpdateTime.String())
		}
		return cat, model.NewError(model.ErrNotFound, id)
	}
	cat.ID = model.MakeCategoryID(dbID)
	return cat, translateError(err)
}

// DeleteCategory deletes a Category
func (p PostgresPersister) DeleteCategory(ctx context.Context, id string) error {
	var dbID int32
	n, err := fmt.Sscanf(id+"\n", p.pathPrefix+model.CategoryIDFormat+"\n", &dbID)
	if err != nil || n != 1 {
		// Bad ID
		return model.NewError(model.ErrNotFound, id)
	}
	_, err = p.db.ExecContext(ctx, "DELETE FROM category WHERE id = $1", dbID)
	return translateError(err)
}
