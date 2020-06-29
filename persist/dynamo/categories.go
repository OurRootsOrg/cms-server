package dynamo

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/ourrootsorg/cms-server/model"
)

const categoryType = "category"

// SelectCategories loads all the categories from the database
func (p Persister) SelectCategories(ctx context.Context) ([]model.Category, error) {
	qi := &dynamodb.QueryInput{
		TableName:              p.tableName,
		IndexName:              aws.String(gsiName),
		KeyConditionExpression: aws.String(skName + " = :sk"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":sk": {
				S: aws.String(categoryType),
			},
		},
	}
	cats := make([]model.Category, 0)
	qo, err := p.svc.Query(qi)
	if err != nil {
		log.Printf("[ERROR] Failed to get categories. qi: %#v err: %v", qi, err)
		return cats, translateError(err)
	}
	err = dynamodbattribute.UnmarshalListOfMaps(qo.Items, &cats)
	if err != nil {
		log.Printf("[ERROR] Failed to unmarshal categories. qo: %#v err: %v", qo, err)
		return cats, translateError(err)
	}
	return cats, nil
}

// SelectCategoriesByID selects many categories
func (p Persister) SelectCategoriesByID(ctx context.Context, ids []uint32) ([]model.Category, error) {
	cats := make([]model.Category, 0)
	if len(ids) == 0 {
		return cats, nil
	}
	keys := make([]map[string]*dynamodb.AttributeValue, len(ids))
	for i, id := range ids {
		keys[i] = map[string]*dynamodb.AttributeValue{
			pkName: {
				N: aws.String(strconv.FormatInt(int64(id), 10)),
			},
			skName: {
				S: aws.String(categoryType),
			},
		}
	}
	bgii := &dynamodb.BatchGetItemInput{
		RequestItems: map[string]*dynamodb.KeysAndAttributes{
			*p.tableName: {
				Keys: keys,
			},
		},
	}
	bgio, err := p.svc.BatchGetItem(bgii)
	if err != nil {
		log.Printf("[ERROR] Failed to get categories. bgii: %#v err: %v", bgii, err)
		return cats, translateError(err)
	}
	err = dynamodbattribute.UnmarshalListOfMaps(bgio.Responses[*p.tableName], &cats)
	if err != nil {
		log.Printf("[ERROR] Failed to unmarshal. bgio: %#v err: %v", bgio, err)
		return cats, translateError(err)
	}
	return cats, nil
}

// SelectOneCategory loads a single category from the database
func (p Persister) SelectOneCategory(ctx context.Context, id uint32) (model.Category, error) {
	var cat model.Category
	gii := &dynamodb.GetItemInput{
		TableName: p.tableName,
		Key: map[string]*dynamodb.AttributeValue{
			pkName: {
				N: aws.String(strconv.FormatInt(int64(id), 10)),
			},
			skName: {
				S: aws.String(categoryType),
			},
		},
	}
	gio, err := p.svc.GetItem(gii)
	if err != nil {
		log.Printf("[ERROR] Failed to get category. qi: %#v err: %v", gio, err)
		return cat, translateError(err)
	}
	if gio.Item == nil {
		return cat, model.NewError(model.ErrNotFound, strconv.Itoa(int(id)))
	}
	err = dynamodbattribute.UnmarshalMap(gio.Item, &cat)
	if err != nil {
		log.Printf("[ERROR] Failed to unmarshal. qo: %#v err: %v", gio, err)
		return cat, translateError(err)
	}
	return cat, nil
}

// InsertCategory inserts a CategoryBody into the database and returns the inserted Category
func (p Persister) InsertCategory(ctx context.Context, in model.CategoryIn) (model.Category, error) {
	var cat model.Category
	var err error
	cat.ID, err = p.getSequenceValue()
	if err != nil {
		return cat, translateError(err)
	}
	cat.Type = categoryType
	cat.CategoryBody = in.CategoryBody
	now := time.Now().Truncate(0)
	cat.InsertTime = now
	cat.LastUpdateTime = now

	avs, err := dynamodbattribute.MarshalMap(cat)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal category %#v: %v", cat, err)
		return cat, translateError(err)
	}

	pii := &dynamodb.PutItemInput{
		TableName:           p.tableName,
		Item:                avs,
		ConditionExpression: aws.String("attribute_not_exists(id)"), // Make duplicate insert fail
	}
	pio, err := p.svc.PutItem(pii)
	if err != nil {
		if compareToAWSError(err, dynamodb.ErrCodeConditionalCheckFailedException) {
			return cat, model.NewError(model.ErrOther, fmt.Sprintf("Insert failed. Category ID %d already exists", cat.ID))
		}
		log.Printf("[ERROR] Failed to put category %#v. pii: %#v err: %v", cat, pii, err)
		return cat, translateError(err)
	}
	err = dynamodbattribute.UnmarshalMap(pio.Attributes, &cat)
	if err != nil {
		return cat, translateError(err)
	}
	return cat, nil
}

// UpdateCategory updates a Category in the database and returns the updated Category
func (p Persister) UpdateCategory(ctx context.Context, id uint32, in model.Category) (model.Category, error) {
	var cat model.Category
	var err error
	cat = in
	cat.ID = id
	cat.Type = categoryType
	cat.LastUpdateTime = time.Now().Truncate(0)
	in.ID = id

	avs, err := dynamodbattribute.MarshalMap(cat)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal category %#v: %v", cat, err)
		return cat, translateError(err)
	}
	lastUpdateTime := in.LastUpdateTime.Format(time.RFC3339Nano)
	pii := &dynamodb.PutItemInput{
		TableName:           p.tableName,
		Item:                avs,
		ConditionExpression: aws.String("last_update_time = :lastUpdateTime"), // Only allow updates, and use optimistic concurrency
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":lastUpdateTime": {S: &lastUpdateTime},
		},
	}
	pio, err := p.svc.PutItem(pii)
	if err != nil {
		if compareToAWSError(err, dynamodb.ErrCodeConditionalCheckFailedException) {
			// Try to retrieve this category. If it doesn't exist, that's why the condition failed
			c, err2 := p.SelectOneCategory(ctx, in.ID)
			if err2 != nil {
				return cat, err2
			}
			// If it does exist, use the LastUpdateTime in the error message
			return cat, model.NewError(model.ErrConcurrentUpdate, c.LastUpdateTime.Format(time.RFC3339Nano), lastUpdateTime)
		}
		log.Printf("[ERROR] Failed to update category %#v. pii: %#v err: %v", cat, pii, err)
		return cat, translateError(err)
	}
	err = dynamodbattribute.UnmarshalMap(pio.Attributes, &cat)
	if err != nil {
		return cat, translateError(err)
	}
	// err := p.db.QueryRowContext(ctx, "UPDATE category SET body = $1, last_update_time = CURRENT_TIMESTAMP WHERE id = $2 AND last_update_time = $3 RETURNING id, body, insert_time, last_update_time", in.CategoryBody, id, in.LastUpdateTime).
	// 	Scan(
	// 		&cat.ID,
	// 		&cat.CategoryBody,
	// 		&cat.InsertTime,
	// 		&cat.LastUpdateTime,
	// 	)
	// if err != nil && err == sql.ErrNoRows {
	// 	// Either non-existent or last_update_time didn't match
	// 	c, _ := p.SelectOneCategory(ctx, id)
	// 	if c.ID == id {
	// 		// Row exists, so it must be a non-matching update time
	// 		return cat, model.NewError(model.ErrConcurrentUpdate, c.LastUpdateTime.String(), in.LastUpdateTime.String())
	// 	}
	// 	return cat, model.NewError(model.ErrNotFound, strconv.Itoa(int(id)))
	// }
	return cat, nil //translateError(err)
}

// DeleteCategory deletes a Category
func (p Persister) DeleteCategory(ctx context.Context, id uint32) error {
	dii := &dynamodb.DeleteItemInput{
		TableName: p.tableName,
		Key: map[string]*dynamodb.AttributeValue{
			pkName: {N: aws.String(strconv.FormatInt(int64(id), 10))},
			skName: {S: aws.String(categoryType)},
		},
	}
	_, err := p.svc.DeleteItem(dii)
	if err != nil {
		return translateError(err)
	}
	return nil
}
