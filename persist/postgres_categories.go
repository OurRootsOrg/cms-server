package persist

import (
	"context"
	"database/sql"
	"log"
	"strconv"

	"github.com/lib/pq"

	"github.com/ourrootsorg/cms-server/model"
)

// SelectCategories loads all the categories from the database
func (p PostgresPersister) SelectCategories(ctx context.Context) ([]model.Category, error) {
	rows, err := p.db.QueryContext(ctx, "SELECT id, body, insert_time, last_update_time FROM category")
	if err != nil {
		return nil, translateError(err, nil, nil, "")
	}
	defer rows.Close()
	cats := make([]model.Category, 0)
	for rows.Next() {
		var cat model.Category
		err := rows.Scan(&cat.ID, &cat.CategoryBody, &cat.InsertTime, &cat.LastUpdateTime)
		if err != nil {
			return nil, translateError(err, nil, nil, "")
		}
		cats = append(cats, cat)
	}
	return cats, nil
}

// SelectCategoriesByID selects many categories
func (p PostgresPersister) SelectCategoriesByID(ctx context.Context, ids []uint32) ([]model.Category, error) {
	categories := make([]model.Category, 0)
	if len(ids) == 0 {
		return categories, nil
	}

	rows, err := p.db.QueryContext(ctx, "SELECT id, body, insert_time, last_update_time FROM category WHERE id = ANY($1)", pq.Array(ids))
	if err != nil {
		return nil, translateError(err, nil, nil, "")
	}
	defer rows.Close()
	for rows.Next() {
		var category model.Category
		err := rows.Scan(&category.ID, &category.CategoryBody, &category.InsertTime, &category.LastUpdateTime)
		if err != nil {
			return nil, translateError(err, nil, nil, "")
		}
		categories = append(categories, category)
	}
	return categories, nil
}

// SelectOneCategory loads a single category from the database
func (p PostgresPersister) SelectOneCategory(ctx context.Context, id uint32) (*model.Category, error) {
	var cat model.Category
	log.Printf("[DEBUG] id: %d", id)
	err := p.db.QueryRowContext(ctx, "SELECT id, body, insert_time, last_update_time FROM category WHERE id=$1", id).Scan(
		&cat.ID,
		&cat.CategoryBody,
		&cat.InsertTime,
		&cat.LastUpdateTime,
	)
	if err != nil {
		return nil, translateError(err, &id, nil, "")
	}
	return &cat, nil
}

// InsertCategory inserts a CategoryBody into the database and returns the inserted Category
func (p PostgresPersister) InsertCategory(ctx context.Context, in model.CategoryIn) (*model.Category, error) {
	var cat model.Category
	row := p.db.QueryRowContext(ctx, "INSERT INTO category (body) VALUES ($1) RETURNING id, body, insert_time, last_update_time", in)
	err := row.Scan(
		&cat.ID,
		&cat.CategoryBody,
		&cat.InsertTime,
		&cat.LastUpdateTime,
	)
	return &cat, translateError(err, nil, nil, "")
}

// UpdateCategory updates a Category in the database and returns the updated Category
func (p PostgresPersister) UpdateCategory(ctx context.Context, id uint32, in model.Category) (*model.Category, error) {
	var cat model.Category
	err := p.db.QueryRowContext(ctx, "UPDATE category SET body = $1, last_update_time = CURRENT_TIMESTAMP WHERE id = $2 AND last_update_time = $3 RETURNING id, body, insert_time, last_update_time", in.CategoryBody, id, in.LastUpdateTime).
		Scan(
			&cat.ID,
			&cat.CategoryBody,
			&cat.InsertTime,
			&cat.LastUpdateTime,
		)
	if err != nil && err == sql.ErrNoRows {
		// Either non-existent or last_update_time didn't match
		c, _ := p.SelectOneCategory(ctx, id)
		if c != nil && c.ID == id {
			// Row exists, so it must be a non-matching update time
			return nil, model.NewError(model.ErrConcurrentUpdate, c.LastUpdateTime.String(), in.LastUpdateTime.String())
		}
		return nil, model.NewError(model.ErrNotFound, strconv.Itoa(int(id)))
	}
	return &cat, translateError(err, &id, nil, "")
}

// DeleteCategory deletes a Category
func (p PostgresPersister) DeleteCategory(ctx context.Context, id uint32) error {
	_, err := p.db.ExecContext(ctx, "DELETE FROM category WHERE id = $1", id)
	return translateError(err, &id, nil, "")
}
