package persist

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/lib/pq"
	"github.com/ourrootsorg/cms-server/model"
)

// PostgresPersister persists the model objects to Postgresql
type PostgresPersister struct {
	pathPrefix string
	db         *sql.DB
}

// NewPostgresPersister constructs a new PostgresPersister
func NewPostgresPersister(pathPrefix string, db *sql.DB) PostgresPersister {
	return PostgresPersister{
		pathPrefix: pathPrefix,
		db:         db,
	}
}

func translateError(err error) error {
	switch err {
	case nil:
		return nil
	case sql.ErrConnDone:
		return ErrConnDone
	case sql.ErrNoRows:
		return ErrNoRows
	case sql.ErrTxDone:
		return ErrTxDone
	default:
		pqErr, ok := err.(*pq.Error)
		if !ok {
			log.Printf("[INFO] Untranslated error: %#v", err)
			return err
		}
		switch pqErr.Code.Name() {
		case "foreign_key_violation":
			return ErrForeignKeyViolation
		default:
			log.Printf("[INFO] Untranslated PQ error: %#v", err)
			return err
		}
	}
}

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
	n, err := fmt.Sscanf(in.Category+"\n", p.pathPrefix+model.CategoryIDFormat+"\n", &catID)
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
	n, err = fmt.Sscanf(in.Category+"\n", p.pathPrefix+model.CategoryIDFormat+"\n", &catID)
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

// Post persistence methods

// SelectPosts selects all posts
func (p PostgresPersister) SelectPosts(ctx context.Context) ([]model.Post, error) {
	rows, err := p.db.QueryContext(ctx, "SELECT id, collection_id, body, insert_time, last_update_time FROM post")
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()
	posts := make([]model.Post, 0)
	for rows.Next() {
		var dbID int32
		var collectionID int32
		var post model.Post
		err := rows.Scan(&dbID, &collectionID, &post.PostBody, &post.InsertTime, &post.LastUpdateTime)
		if err != nil {
			return nil, translateError(err)
		}
		post.ID = model.MakePostID(dbID)
		post.Collection = model.MakeCollectionID(collectionID)
		posts = append(posts, post)
	}
	return posts, nil
}

// SelectOnePost selects a single post
func (p PostgresPersister) SelectOnePost(ctx context.Context, id string) (model.Post, error) {
	var post model.Post
	var dbID int32
	var catID int32
	n, err := fmt.Sscanf(id+"\n", p.pathPrefix+model.PostIDFormat+"\n", &dbID)
	if err != nil || n != 1 {
		// Bad ID
		return post, model.NewError(model.ErrNotFound, id)
	}
	err = p.db.QueryRowContext(ctx, "SELECT id, collection_id, body, insert_time, last_update_time FROM post WHERE id=$1", dbID).Scan(
		&dbID,
		&catID,
		&post.PostBody,
		&post.InsertTime,
		&post.LastUpdateTime,
	)
	if err != nil {
		return post, translateError(err)
	}
	post.ID = model.MakePostID(dbID)
	post.Collection = model.MakeCollectionID(catID)
	return post, nil
}

// InsertPost inserts a PostBody into the database and returns the inserted Post
func (p PostgresPersister) InsertPost(ctx context.Context, in model.PostIn) (model.Post, error) {
	var dbID int32
	var post model.Post
	var catID int32
	n, err := fmt.Sscanf(in.Collection+"\n", p.pathPrefix+model.CollectionIDFormat+"\n", &catID)
	if err != nil || n != 1 {
		// Bad ID
		return post, model.NewError(model.ErrBadReference, in.Collection, "collection")
	}
	err = p.db.QueryRowContext(ctx,
		`INSERT INTO post (collection_id, body) 
		 VALUES ($1, $2) 
		 RETURNING id, collection_id, body, insert_time, last_update_time`,
		catID, in.PostBody).
		Scan(
			&dbID,
			&catID,
			&post.PostBody,
			&post.InsertTime,
			&post.LastUpdateTime,
		)
	post.ID = model.MakePostID(dbID)
	post.Collection = model.MakeCollectionID(catID)
	return post, translateError(err)
}

// UpdatePost updates a Post in the database and returns the updated Post
func (p PostgresPersister) UpdatePost(ctx context.Context, id string, in model.Post) (model.Post, error) {
	var post model.Post
	var dbID int32
	n, err := fmt.Sscanf(id+"\n", p.pathPrefix+model.PostIDFormat+"\n", &dbID)
	if err != nil || n != 1 {
		// Bad ID
		return post, model.NewError(model.ErrNotFound, id)
	}
	var catID int32
	n, err = fmt.Sscanf(in.Collection+"\n", p.pathPrefix+model.CollectionIDFormat+"\n", &catID)
	if err != nil || n != 1 {
		// Bad ID
		return post, model.NewError(model.ErrBadReference, in.Collection, "collection")
	}
	err = p.db.QueryRowContext(ctx,
		`UPDATE post SET body = $1, collection_id = $2, last_update_time = CURRENT_TIMESTAMP 
		 WHERE id = $3 AND last_update_time = $4
		 RETURNING id, collection_id, body, insert_time, last_update_time`,
		in.PostBody, catID, dbID, in.LastUpdateTime).
		Scan(
			&dbID,
			&catID,
			&post.PostBody,
			&post.InsertTime,
			&post.LastUpdateTime,
		)
	if err != nil && err == sql.ErrNoRows {
		// Either non-existent or last_update_time didn't match
		c, _ := p.SelectOnePost(ctx, id)
		if c.ID == id {
			// Row exists, so it must be a non-matching update time
			return post, model.NewError(model.ErrConcurrentUpdate, c.LastUpdateTime.String(), in.LastUpdateTime.String())
		}
		return post, model.NewError(model.ErrNotFound, id)
	}
	post.ID = model.MakePostID(dbID)
	post.Collection = model.MakeCollectionID(catID)
	return post, translateError(err)
}

// DeletePost deletes a Post
func (p PostgresPersister) DeletePost(ctx context.Context, id string) error {
	var dbID int32
	n, err := fmt.Sscanf(id+"\n", p.pathPrefix+model.PostIDFormat+"\n", &dbID)
	if err != nil || n != 1 {
		// Bad ID
		return model.NewError(model.ErrNotFound, id)
	}
	_, err = p.db.ExecContext(ctx, "DELETE FROM post WHERE id = $1", dbID)
	return translateError(err)
}
