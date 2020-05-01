package persist

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/lib/pq"
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
	case sql.ErrConnDone:
		return ErrConnDone
	case sql.ErrNoRows:
		return ErrNoRows
	case sql.ErrTxDone:
		return ErrTxDone
	case nil:
		return nil
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
func (p PostgresPersister) SelectCategories() ([]model.Category, error) {
	rows, err := p.db.Query("SELECT id, body, insert_time, last_update_time FROM category")
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()
	cats := make([]model.Category, 0)
	for rows.Next() {
		var dbid int32
		var cat model.Category
		err := rows.Scan(&dbid, &cat.CategoryBody, &cat.InsertTime, &cat.LastUpdateTime)
		if err != nil {
			return nil, translateError(err)
		}
		cat.ID = p.pathPrefix + fmt.Sprintf(model.CategoryIDFormat, dbid)
		cat.Type = "category"
		cats = append(cats, cat)
	}
	return cats, nil
}

// SelectOneCategory loads a single category from the database
func (p PostgresPersister) SelectOneCategory(id string) (model.Category, error) {
	var dbid int32
	fmt.Sscanf(id, p.pathPrefix+model.CategoryIDFormat, &dbid)
	var cat model.Category
	log.Printf("[DEBUG] id: %s, dbid: %d", id, dbid)
	err := p.db.QueryRow("SELECT id, body, insert_time, last_update_time FROM category WHERE id=$1", dbid).Scan(
		&dbid,
		&cat.CategoryBody,
		&cat.InsertTime,
		&cat.LastUpdateTime,
	)
	if err != nil {
		return cat, translateError(err)
	}
	cat.ID = p.pathPrefix + fmt.Sprintf(model.CategoryIDFormat, dbid)
	cat.Type = "category"
	return cat, nil
}

// InsertCategory inserts a CategoryBody into the database and returns the inserted Category
func (p PostgresPersister) InsertCategory(in model.CategoryIn) (model.Category, error) {
	var dbid int32
	var cat model.Category
	err := p.db.QueryRow("INSERT INTO category (body) VALUES ($1) RETURNING id, body, insert_time, last_update_time", in).
		Scan(
			&dbid,
			&cat.CategoryBody,
			&cat.InsertTime,
			&cat.LastUpdateTime,
		)
	cat.ID = p.pathPrefix + fmt.Sprintf(model.CategoryIDFormat, dbid)
	cat.Type = "category"
	return cat, translateError(err)
}

// UpdateCategory updates a Category in the database and returns the updated Category
func (p PostgresPersister) UpdateCategory(id string, in model.CategoryIn) (model.Category, error) {
	var dbid int32
	fmt.Sscanf(id, p.pathPrefix+model.CategoryIDFormat, &dbid)
	var cat model.Category
	err := p.db.QueryRow("UPDATE category SET body = $1, last_update_time = CURRENT_TIMESTAMP WHERE id = $2 RETURNING id, body, insert_time, last_update_time", in, dbid).
		Scan(
			&dbid,
			&cat.CategoryBody,
			&cat.InsertTime,
			&cat.LastUpdateTime,
		)
	cat.ID = p.pathPrefix + fmt.Sprintf(model.CategoryIDFormat, dbid)
	cat.Type = "category"
	return cat, translateError(err)
}

// DeleteCategory deletes a Category
func (p PostgresPersister) DeleteCategory(id string) error {
	var dbid int32
	fmt.Sscanf(id, p.pathPrefix+model.CategoryIDFormat, &dbid)
	_, err := p.db.Exec("DELETE FROM category WHERE id = $1", dbid)
	return translateError(err)
}

// Collection persistence methods

// SelectCollections selects all collections
func (p PostgresPersister) SelectCollections() ([]model.Collection, error) {
	rows, err := p.db.Query("SELECT id, category_id, body, insert_time, last_update_time FROM collection")
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()
	collections := make([]model.Collection, 0)
	for rows.Next() {
		var dbid int32
		var categoryID int32
		var collection model.Collection
		err := rows.Scan(&dbid, &categoryID, &collection.CollectionBody, &collection.InsertTime, &collection.LastUpdateTime)
		if err != nil {
			return nil, translateError(err)
		}
		collection.ID = p.pathPrefix + fmt.Sprintf(model.CollectionIDFormat, dbid)
		collection.Category.ID = p.pathPrefix + fmt.Sprintf(model.CategoryIDFormat, categoryID)
		collection.Category.Type = "category"
		collection.Type = "collection"
		collections = append(collections, collection)
	}
	return collections, nil
}

// SelectOneCollection selects a single collection
func (p PostgresPersister) SelectOneCollection(id string) (model.Collection, error) {
	var dbid int32
	fmt.Sscanf(id, p.pathPrefix+model.CollectionIDFormat, &dbid)
	var collection model.Collection
	err := p.db.QueryRow("SELECT id, body, insert_time, last_update_time FROM collection WHERE id=$1", dbid).Scan(
		&dbid,
		&collection.CollectionBody,
		&collection.InsertTime,
		&collection.LastUpdateTime,
	)
	if err != nil {
		return collection, translateError(err)
	}
	collection.ID = p.pathPrefix + fmt.Sprintf(model.CollectionIDFormat, dbid)
	collection.Type = "collection"
	return collection, nil
}

// InsertCollection inserts a CollectionBody into the database and returns the inserted Collection
func (p PostgresPersister) InsertCollection(in model.CollectionIn) (model.Collection, error) {
	var dbid int32
	var collection model.Collection
	err := p.db.QueryRow(
		`INSERT INTO collection (category_id, body) 
		 VALUES ($1, $2) 
		 RETURNING id, category_id, body, insert_time, last_update_time`,
		in.Category, in.CollectionBody).
		Scan(
			&dbid,
			&collection.Category,
			&collection.CollectionBody,
			&collection.InsertTime,
			&collection.LastUpdateTime,
		)
	collection.ID = p.pathPrefix + fmt.Sprintf(model.CollectionIDFormat, dbid)
	collection.Type = "collection"
	return collection, translateError(err)
}

// UpdateCollection updates a Collection in the database and returns the updated Collection
func (p PostgresPersister) UpdateCollection(id string, in model.CollectionIn) (model.Collection, error) {
	var dbid int32
	fmt.Sscanf(id, p.pathPrefix+model.CollectionIDFormat, &dbid)
	var catID int32
	fmt.Sscanf(in.Category.ID, p.pathPrefix+model.CategoryIDFormat, &catID)
	var collection model.Collection
	err := p.db.QueryRow(
		`UPDATE collection SET body = $1, category_id = $2, last_update_time = CURRENT_TIMESTAMP 
		 WHERE id = $3 
		 RETURNING id, category_id, body, insert_time, last_update_time`,
		in.CollectionBody, catID, dbid).
		Scan(
			&dbid,
			&collection.Category,
			&collection.CollectionBody,
			&collection.InsertTime,
			&collection.LastUpdateTime,
		)
	collection.ID = p.pathPrefix + fmt.Sprintf(model.CollectionIDFormat, dbid)
	collection.Type = "collection"
	return collection, translateError(err)
}

// DeleteCollection deletes a Collection
func (p PostgresPersister) DeleteCollection(id string) error {
	var dbid int32
	fmt.Sscanf(id, p.pathPrefix+model.CollectionIDFormat, &dbid)
	_, err := p.db.Exec("DELETE FROM collection WHERE id = $1", dbid)
	return translateError(err)
}
