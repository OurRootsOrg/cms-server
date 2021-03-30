package persist_test

import (
	"context"
	"encoding/json"
	"log"
	"testing"
	"time"

	"github.com/ourrootsorg/cms-server/utils"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestSelectCategories(t *testing.T) {
	ctx := utils.AddSocietyIDToContext(context.TODO(), 1)
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()
	p := persist.NewPostgresPersister(db)
	in := makeCategoryIn(t)
	js, err := json.Marshal(in)
	assert.NoError(t, err)

	now := time.Now()

	mock.ExpectQuery("SELECT id, body, insert_time, last_update_time FROM category WHERE society_id=$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "body", "insert_time", "last_update_time"}).
			AddRow(1, js, now, now).AddRow(2, js, now, now))

	c, e := p.SelectCategories(ctx)
	assert.Nil(t, e)
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
	ctx := utils.AddSocietyIDToContext(context.TODO(), 1)
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()
	p := persist.NewPostgresPersister(db)

	cb := makeCategoryIn(t)
	js, err := json.Marshal(cb)
	assert.NoError(t, err)

	now := time.Now()
	mock.ExpectQuery("SELECT id, body, insert_time, last_update_time FROM category WHERE society_id=$1 AND id=$2").
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "body", "insert_time", "last_update_time"}).
			AddRow(int32(1), js, now, now))

	c, e := p.SelectOneCategory(ctx, 1)
	assert.Nil(t, e)
	assert.Equal(t, uint32(1), c.ID)
	assert.Equal(t, cb.Name, c.Name)
	assert.Equal(t, now, c.InsertTime)
	assert.Equal(t, now, c.LastUpdateTime)
}

func TestInsertCategory(t *testing.T) {
	ctx := utils.AddSocietyIDToContext(context.TODO(), 1)
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()
	p := persist.NewPostgresPersister(db)
	cb := makeCategoryIn(t)
	js, err := json.Marshal(cb)
	assert.NoError(t, err)

	now := time.Now()
	mock.ExpectQuery("INSERT INTO category (society_id, body) VALUES ($1,$2) RETURNING id, body, insert_time, last_update_time").
		WithArgs(1, []byte(js)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "body", "insert_time", "last_update_time"}).AddRow(1, js, now, now))

	c, e := p.InsertCategory(ctx, cb)
	assert.Nil(t, e)
	assert.Equal(t, uint32(1), c.ID)
	assert.Equal(t, cb.Name, c.Name)
	assert.Equal(t, now, c.InsertTime)
	assert.Equal(t, now, c.LastUpdateTime)
}

func TestUpdateCategory(t *testing.T) {
	ctx := utils.AddSocietyIDToContext(context.TODO(), 1)
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()
	p := persist.NewPostgresPersister(db)
	in := makeCategory(t)
	js, err := json.Marshal(in.CategoryBody)
	assert.NoError(t, err)

	now := time.Now()
	mock.ExpectQuery("UPDATE category SET body = $1, last_update_time = CURRENT_TIMESTAMP WHERE society_id = $2 AND id = $3 AND last_update_time = $4 RETURNING id, body, insert_time, last_update_time").
		WithArgs([]byte(js), 1, 1, in.LastUpdateTime).
		WillReturnRows(sqlmock.NewRows([]string{"id", "body", "insert_time", "last_update_time"}).AddRow(1, js, in.InsertTime, now))

	c, e := p.UpdateCategory(ctx, 1, in)
	assert.Nil(t, e)
	assert.Equal(t, uint32(1), c.ID)
	assert.Equal(t, in.Name, c.Name)
	assert.Equal(t, in.InsertTime, c.InsertTime)
	assert.Equal(t, now, c.LastUpdateTime)
}
func TestDeleteCategory(t *testing.T) {
	ctx := utils.AddSocietyIDToContext(context.TODO(), 1)
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()
	p := persist.NewPostgresPersister(db)
	mock.ExpectExec("DELETE FROM category WHERE society_id = $1 AND id = $2").
		WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(0, 1))
	e := p.DeleteCategory(ctx, 1)
	assert.Nil(t, e)
}

// Collection tests

func TestSelectCollections(t *testing.T) {
	ctx := utils.AddSocietyIDToContext(context.TODO(), 1)
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()
	p := persist.NewPostgresPersister(db)
	in := makeCollectionIn(t)
	js, err := json.Marshal(in.CollectionBody)
	assert.NoError(t, err)
	log.Printf("[DEBUG] json: %s", string(js))

	now := time.Now()

	mock.ExpectQuery(
		"SELECT id, array_agg(cc.category_id), body, insert_time, last_update_time " +
			"FROM collection LEFT JOIN collection_category cc ON id = cc.collection_id WHERE society_id=$1 GROUP BY id").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "category_id", "body", "insert_time", "last_update_time"}).
			AddRow(1, "{1}", js, now, now).
			AddRow(2, "{1}", js, now, now))

	c, e := p.SelectCollections(ctx)
	assert.Nil(t, e)
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
	ctx := utils.AddSocietyIDToContext(context.TODO(), 1)
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()
	p := persist.NewPostgresPersister(db)

	cb := makeCollectionIn(t)
	js, err := json.Marshal(cb)
	assert.NoError(t, err)

	now := time.Now()
	mock.ExpectQuery(
		"SELECT id, array_agg(cc.category_id), body, insert_time, last_update_time "+
			"FROM collection LEFT JOIN collection_category cc ON id = cc.collection_id WHERE society_id=$1 AND id = $2 GROUP BY id").
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "category_id", "body", "insert_time", "last_update_time"}).
			AddRow(1, "{1}", js, now, now))

	c, e := p.SelectOneCollection(ctx, 1)
	assert.Nil(t, e)
	assert.Equal(t, uint32(1), c.ID)
	assert.Equal(t, cb.Name, c.Name)
	assert.Equal(t, now, c.InsertTime)
	assert.Equal(t, now, c.LastUpdateTime)
}

func TestInsertCollection(t *testing.T) {
	ctx := utils.AddSocietyIDToContext(context.TODO(), 1)
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()
	p := persist.NewPostgresPersister(db)
	in := makeCollectionIn(t)
	js, err := json.Marshal(in.CollectionBody)
	assert.NoError(t, err)

	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO collection (society_id, body)
	VALUES ($1, $2)
	RETURNING id, body, insert_time, last_update_time`).
		WithArgs(1, []byte(js)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "body", "insert_time", "last_update_time"}).
			AddRow(1, js, now, now))
	mock.ExpectExec("INSERT INTO collection_category (collection_id, category_id) VALUES ($1, $2)").
		WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	c, e := p.InsertCollection(ctx, in)
	assert.Nil(t, e)
	assert.Equal(t, uint32(1), c.ID)
	assert.Equal(t, in.Name, c.Name)
	assert.Equal(t, now, c.InsertTime)
	assert.Equal(t, now, c.LastUpdateTime)
}

func TestUpdateCollection(t *testing.T) {
	ctx := utils.AddSocietyIDToContext(context.TODO(), 1)
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()
	p := persist.NewPostgresPersister(db)
	collExist := makeCollectionIn(t)
	jsExist, err := json.Marshal(collExist.CollectionBody)
	assert.NoError(t, err)
	in := makeCollection(t)
	js, err := json.Marshal(in.CollectionBody)
	assert.NoError(t, err)

	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(
		"SELECT id, array_agg(cc.category_id), body, insert_time, last_update_time "+
			"FROM collection LEFT JOIN collection_category cc ON id = cc.collection_id WHERE society_id=$1 AND id = $2 GROUP BY id").
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "category_id", "body", "insert_time", "last_update_time"}).
			AddRow(1, "{1}", jsExist, now, now))
	mock.ExpectQuery(`UPDATE collection SET body = $1, last_update_time = CURRENT_TIMESTAMP
	WHERE id = $2 AND last_update_time = $3
	RETURNING id, body, insert_time, last_update_time`).
		WithArgs([]byte(js), 1, in.LastUpdateTime).
		WillReturnRows(sqlmock.NewRows([]string{"id", "body", "insert_time", "last_update_time"}).
			AddRow(1, js, now, now))
	mock.ExpectExec("DELETE FROM collection_category WHERE collection_id = $1").
		WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO collection_category (collection_id, category_id) VALUES ($1, $2)").
		WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	c, e := p.UpdateCollection(ctx, 1, in)
	assert.Nil(t, e)
	assert.Equal(t, uint32(1), c.ID)
	assert.Equal(t, in.Name, c.Name)
	assert.Equal(t, now, c.InsertTime)
	assert.Equal(t, now, c.LastUpdateTime)
}
func TestDeleteCollection(t *testing.T) {
	ctx := utils.AddSocietyIDToContext(context.TODO(), 1)
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()
	p := persist.NewPostgresPersister(db)

	collExist := makeCollectionIn(t)
	jsExist, err := json.Marshal(collExist)
	assert.NoError(t, err)
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(
		"SELECT id, array_agg(cc.category_id), body, insert_time, last_update_time "+
			"FROM collection LEFT JOIN collection_category cc ON id = cc.collection_id WHERE society_id=$1 AND id = $2 GROUP BY id").
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "category_id", "body", "insert_time", "last_update_time"}).
			AddRow(1, "{1}", jsExist, now, now))
	mock.ExpectExec("DELETE FROM collection_category WHERE collection_id = $1").
		WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM collection WHERE id = $1").
		WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	e := p.DeleteCollection(ctx, 1)
	assert.Nil(t, e)
}

func makeCategoryIn(t *testing.T) model.CategoryIn {
	in, e := model.NewCategoryIn("Test Category")
	assert.Nil(t, e)
	return in
}
func makeCategory(t *testing.T) model.Category {
	now := time.Now()
	in := model.Category{
		ID:             33,
		CategoryBody:   makeCategoryIn(t).CategoryBody,
		InsertTime:     now,
		LastUpdateTime: now,
	}
	return in
}

func makeCollectionIn(t *testing.T) model.CollectionIn {
	in := model.CollectionIn{
		CollectionBody: model.CollectionBody{
			Name:           "Test Collection",
			CollectionType: model.CollectionTypeRecords,
		},
		Categories: []uint32{1},
	}
	return in
}

func makeCollection(t *testing.T) model.Collection {
	now := time.Now()
	in := model.Collection{
		ID:             22,
		CollectionIn:   makeCollectionIn(t),
		InsertTime:     now,
		LastUpdateTime: now,
	}
	return in
}
