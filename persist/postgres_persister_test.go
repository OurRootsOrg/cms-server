package persist_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestSelectCategories(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()
	p := persist.NewPostgresPersister("", db)
	in := makeCategoryIn(t)
	js, err := json.Marshal(in)
	assert.NoError(t, err)

	now := time.Now()

	mock.ExpectQuery("SELECT id, body, insert_time, last_update_time FROM category").
		WillReturnRows(sqlmock.NewRows([]string{"id", "body", "insert_time", "last_update_time"}).
			AddRow(1, js, now, now).AddRow(2, js, now, now))

	c, err := p.SelectCategories(context.TODO())
	assert.NoError(t, err)
	assert.Len(t, c, 2)
	cc := model.NewCategory(1, in)
	cc.InsertTime = now
	cc.LastUpdateTime = now
	assert.Contains(t, c, cc)
	cc = model.NewCategory(1, in)
	cc.InsertTime = now
	cc.LastUpdateTime = now
	assert.Contains(t, c, cc)
}
func TestSelectOneCategory(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()
	p := persist.NewPostgresPersister("", db)

	cb := makeCategoryIn(t)
	js, err := json.Marshal(cb)
	assert.NoError(t, err)

	now := time.Now()
	mock.ExpectQuery("SELECT id, body, insert_time, last_update_time FROM category WHERE id=$1").
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "body", "insert_time", "last_update_time"}).
			AddRow(int32(1), js, now, now))

	c, err := p.SelectOneCategory(context.TODO(), "/categories/1")
	assert.NoError(t, err)
	assert.Equal(t, "/categories/1", c.ID)
	assert.Equal(t, cb.Name, c.Name)
	assert.Equal(t, cb.FieldDefs, c.FieldDefs)
	assert.Equal(t, now, c.InsertTime)
	assert.Equal(t, now, c.LastUpdateTime)
}

func TestInsertCategory(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()
	p := persist.NewPostgresPersister("", db)
	cb := makeCategoryIn(t)
	js, err := json.Marshal(cb)
	assert.NoError(t, err)

	now := time.Now()
	mock.ExpectQuery("INSERT INTO category (body) VALUES ($1) RETURNING id, body, insert_time, last_update_time").
		WithArgs([]byte(js)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "body", "insert_time", "last_update_time"}).AddRow(1, js, now, now))

	c, err := p.InsertCategory(context.TODO(), cb)
	assert.NoError(t, err)
	assert.Equal(t, "/categories/1", c.ID)
	assert.Equal(t, cb.Name, c.Name)
	assert.Equal(t, cb.FieldDefs, c.FieldDefs)
	assert.Equal(t, now, c.InsertTime)
	assert.Equal(t, now, c.LastUpdateTime)
}

func TestUpdateCategory(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()
	p := persist.NewPostgresPersister("", db)
	in := makeCategory(t)
	js, err := json.Marshal(in.CategoryBody)
	assert.NoError(t, err)

	now := time.Now()
	mock.ExpectQuery("UPDATE category SET body = $1, last_update_time = CURRENT_TIMESTAMP WHERE id = $2 AND last_update_time = $3 RETURNING id, body, insert_time, last_update_time").
		WithArgs([]byte(js), 1, in.LastUpdateTime).
		WillReturnRows(sqlmock.NewRows([]string{"id", "body", "insert_time", "last_update_time"}).AddRow(1, js, in.InsertTime, now))

	c, err := p.UpdateCategory(context.TODO(), "/categories/1", in)
	assert.NoError(t, err)
	assert.Equal(t, "/categories/1", c.ID)
	assert.Equal(t, in.Name, c.Name)
	assert.Equal(t, in.FieldDefs, c.FieldDefs)
	assert.Equal(t, in.InsertTime, c.InsertTime)
	assert.Equal(t, now, c.LastUpdateTime)
}
func TestDeleteCategory(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()
	p := persist.NewPostgresPersister("", db)
	mock.ExpectExec("DELETE FROM category WHERE id = $1").
		WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
	err = p.DeleteCategory(context.TODO(), "/categories/1")
	assert.NoError(t, err)
}

// Collection tests

func TestSelectCollections(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()
	p := persist.NewPostgresPersister("", db)
	in := makeCollectionIn(t)
	js, err := json.Marshal(in.CollectionBody)
	assert.NoError(t, err)
	log.Printf("[DEBUG] json: %s", string(js))

	now := time.Now()

	mock.ExpectQuery("SELECT id, category_id, body, insert_time, last_update_time FROM collection").
		WillReturnRows(sqlmock.NewRows([]string{"id", "category_id", "body", "insert_time", "last_update_time"}).
			AddRow(1, 1, js, now, now).
			AddRow(2, 1, js, now, now))

	c, err := p.SelectCollections(context.TODO())
	assert.NoError(t, err)
	assert.Len(t, c, 2)
	cc := model.NewCollection(1, in)
	cc.InsertTime = now
	cc.LastUpdateTime = now
	assert.Contains(t, c, cc)

	cc = model.NewCollection(2, in)
	cc.InsertTime = now
	cc.LastUpdateTime = now
	assert.Contains(t, c, cc)
}
func TestSelectOneCollection(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()
	p := persist.NewPostgresPersister("", db)

	cb := makeCollectionIn(t)
	js, err := json.Marshal(cb)
	assert.NoError(t, err)

	now := time.Now()
	mock.ExpectQuery("SELECT id, category_id, body, insert_time, last_update_time FROM collection WHERE id=$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "category_id", "body", "insert_time", "last_update_time"}).
			AddRow(1, 1, js, now, now))

	c, err := p.SelectOneCollection(context.TODO(), "/collections/1")
	assert.NoError(t, err)
	assert.Equal(t, "/collections/1", c.ID)
	assert.Equal(t, cb.Name, c.Name)
	assert.Equal(t, now, c.InsertTime)
	assert.Equal(t, now, c.LastUpdateTime)
}

func TestInsertCollection(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()
	p := persist.NewPostgresPersister("", db)
	in := makeCollectionIn(t)
	var catID int32
	fmt.Sscanf(in.Category+"\n", model.CategoryIDFormat, &catID)
	js, err := json.Marshal(in.CollectionBody)
	assert.NoError(t, err)

	now := time.Now()
	mock.ExpectQuery(`INSERT INTO collection (category_id, body) 
	VALUES ($1, $2) 
	RETURNING id, category_id, body, insert_time, last_update_time`).
		WithArgs(catID, []byte(js)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "category_id", "body", "insert_time", "last_update_time"}).
			AddRow(1, catID, js, now, now))

	c, err := p.InsertCollection(context.TODO(), in)
	assert.NoError(t, err)
	assert.Equal(t, "/collections/1", c.ID)
	assert.Equal(t, in.Name, c.Name)
	assert.Equal(t, now, c.InsertTime)
	assert.Equal(t, now, c.LastUpdateTime)
}

func TestUpdateCollection(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()
	p := persist.NewPostgresPersister("", db)
	in := makeCollection(t)
	js, err := json.Marshal(in.CollectionBody)
	assert.NoError(t, err)
	var catID int32
	fmt.Sscanf(in.Category+"\n", model.CategoryIDFormat, &catID)

	now := time.Now()
	mock.ExpectQuery(`UPDATE collection SET body = $1, category_id = $2, last_update_time = CURRENT_TIMESTAMP 
	WHERE id = $3 AND last_update_time = $4
	RETURNING id, category_id, body, insert_time, last_update_time`).
		WithArgs([]byte(js), catID, 1, in.LastUpdateTime).
		WillReturnRows(sqlmock.NewRows([]string{"id", "category_id", "body", "insert_time", "last_update_time"}).
			AddRow(1, 1, js, now, now))

	c, err := p.UpdateCollection(context.TODO(), "/collections/1", in)
	assert.NoError(t, err)
	assert.Equal(t, "/collections/1", c.ID)
	assert.Equal(t, in.Name, c.Name)
	assert.Equal(t, now, c.InsertTime)
	assert.Equal(t, now, c.LastUpdateTime)
}
func TestDeleteCollection(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()
	p := persist.NewPostgresPersister("", db)
	mock.ExpectExec("DELETE FROM collection WHERE id = $1").
		WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
	err = p.DeleteCollection(context.TODO(), "/collections/1")
	assert.NoError(t, err)
}

func makeCategoryIn(t *testing.T) model.CategoryIn {
	intType, err := model.NewFieldDef("intField", model.IntType, "int_field")
	assert.NoError(t, err)
	stringType, err := model.NewFieldDef("stringField", model.StringType, "string_field")
	assert.NoError(t, err)
	imageType, err := model.NewFieldDef("imageField", model.ImageType, "image_field")
	assert.NoError(t, err)
	locationType, err := model.NewFieldDef("locationField", model.LocationType, "location_field")
	assert.NoError(t, err)
	timeType, err := model.NewFieldDef("timeField", model.TimeType, "time_field")
	assert.NoError(t, err)
	in, err := model.NewCategoryIn("Test Category", intType, stringType, imageType, locationType, timeType)
	assert.NoError(t, err)
	return in
}
func makeCategory(t *testing.T) model.Category {
	now := time.Now()
	in := model.Category{
		ID:             model.MakeCategoryID(33),
		CategoryIn:     makeCategoryIn(t),
		InsertTime:     now,
		LastUpdateTime: now,
	}
	return in
}

func makeCollectionIn(t *testing.T) model.CollectionIn {
	in := model.CollectionIn{
		CollectionBody: model.CollectionBody{Name: "Test Collection"},
		Category:       model.MakeCategoryID(1),
	}
	return in
}

func makeCollection(t *testing.T) model.Collection {
	now := time.Now()
	in := model.Collection{
		ID:             model.MakeCollectionID(22),
		CollectionIn:   makeCollectionIn(t),
		InsertTime:     now,
		LastUpdateTime: now,
	}
	return in
}
