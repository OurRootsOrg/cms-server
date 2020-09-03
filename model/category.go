package model

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// CategoryPersister defines methods needed to persist categories
type CategoryPersister interface {
	SelectCategories(ctx context.Context) ([]Category, error)
	SelectCategoriesByID(ctx context.Context, ids []uint32) ([]Category, error)
	SelectOneCategory(ctx context.Context, id uint32) (*Category, error)
	InsertCategory(ctx context.Context, in CategoryIn) (*Category, error)
	UpdateCategory(ctx context.Context, id uint32, body Category) (*Category, error)
	DeleteCategory(ctx context.Context, id uint32) error
}

// CategoryIn is the payload to create or update a category
type CategoryIn struct {
	CategoryBody
}

// CategoryBody is the JSON part of the Category object
type CategoryBody struct {
	Name string `json:"name" validate:"required" dynamodbav:"altSort"`
}

// NewCategoryIn constructs a CategoryIn
func NewCategoryIn(name string) (CategoryIn, error) {
	cb, err := newCategoryBody(name)
	if err != nil {
		return CategoryIn{}, err
	}
	return CategoryIn{CategoryBody: cb}, nil
}

// newCategoryBody constructs a CategoryBody
func newCategoryBody(name string) (CategoryBody, error) {
	cb := CategoryBody{Name: name}
	return cb, nil
}

// Value makes CategoryBody implement the driver.Valuer interface.
func (cb CategoryBody) Value() (driver.Value, error) {
	return json.Marshal(cb)
}

// Scan makes CategoryBody implement the sql.Scanner interface.
func (cb *CategoryBody) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &cb)
}

// Category represents a set of collections
type Category struct {
	ID   uint32 `json:"id,omitempty" example:"999" validate:"required,omitempty" dynamodbav:"pk,string"`
	Type string `json:"-" dynamodbav:"sk"`
	CategoryBody
	InsertTime     time.Time `json:"insert_time,omitempty"`
	LastUpdateTime time.Time `json:"last_update_time,omitempty"`
}

// NewCategory constructs a Category from an id and body
func NewCategory(id uint32, in CategoryIn) Category {
	return Category{
		ID:           id,
		CategoryBody: in.CategoryBody,
	}
}

// func (cat *Category) MarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
// 	avs, err := dynamodbattribute.MarshalMap(cat)
// 	if err != nil {
// 		return err
// 	}
// 	log.Printf("[DEBUG] MarshalDynamoDBAttributeValue cat = %#v, av = %#v", cat, avs)
// 	av.SetM(avs)
// 	av.M["ID"].SetS(*av.M["Type"].S + "#" + *avs["ID"].S)
// 	log.Printf("[DEBUG] MarshalDynamoDBAttributeValue updated ID, cat = %#v, avs = %#v, av = %#v", cat, avs, av)
// 	return nil
// }

// func (cat *Category) UnmarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
// 	cat = &Category{}
// 	err := dynamodbattribute.UnmarshalMap(av.M, cat)
// 	if err != nil {
// 		log.Printf("[ERROR] Error unmarshalling %#v into Category: %v", av, err)
// 		return err
// 	}
// 	log.Printf("[DEBUG] UnmarshalDynamoDBAttributeValue cat = %#v, av = %#v", cat, av)
// 	id, err := strconv.ParseUint(strings.TrimPrefix(*av.M["ID"].S, *av.M["Type"].S+"#"), 10, 32)
// 	if err != nil {
// 		log.Printf("[ERROR] Error calling ParseUint on %s: %v", strings.TrimPrefix(*av.M["ID"].S, *av.M["Type"].S+"#"), err)
// 		return err
// 	}
// 	cat.ID = uint32(id)
// 	return nil
// }
