package dynamo

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/ourrootsorg/cms-server/model"
)

const (
	collectionType           = "collection"
	collectionCategoryPrefix = collectionType + "_" + categoryType + "#"
)

type collectionCategory struct {
	ID uint32 `dynamodbav:"pk"`
	SK string `dynamodbav:"sk"`
}

// SelectCollections loads all the collections from the database
func (p Persister) SelectCollections(ctx context.Context) ([]model.Collection, *model.Error) {
	qi := &dynamodb.QueryInput{
		TableName:              p.tableName,
		IndexName:              aws.String(gsiName),
		KeyConditionExpression: aws.String(skName + " = :sk"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":sk": {
				S: aws.String(collectionType),
			},
		},
	}
	colls := make([]model.Collection, 0)
	qo, err := p.svc.Query(qi)
	if err != nil {
		log.Printf("[ERROR] Failed to get collections. qi: %#v err: %v", qi, err)
		return nil, model.NewError(model.ErrOther, err.Error())
	}
	err = dynamodbattribute.UnmarshalListOfMaps(qo.Items, &colls)
	if err != nil {
		log.Printf("[ERROR] Failed to unmarshal collections. qo: %#v err: %v", qo, err)
		return nil, model.NewError(model.ErrOther, err.Error())
	}
	return colls, nil
}

// SelectCollectionsByID selects many collections
func (p Persister) SelectCollectionsByID(ctx context.Context, ids []uint32) ([]model.Collection, *model.Error) {
	colls := make([]model.Collection, 0)
	if len(ids) == 0 {
		return colls, nil
	}
	keys := make([]map[string]*dynamodb.AttributeValue, len(ids))
	for i, id := range ids {
		keys[i] = map[string]*dynamodb.AttributeValue{
			pkName: {
				N: aws.String(strconv.FormatInt(int64(id), 10)),
			},
			skName: {
				S: aws.String(collectionType),
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
		log.Printf("[ERROR] Failed to get collections. bgii: %#v err: %v", bgii, err)
		return nil, model.NewError(model.ErrOther, err.Error())
	}
	err = dynamodbattribute.UnmarshalListOfMaps(bgio.Responses[*p.tableName], &colls)
	if err != nil {
		log.Printf("[ERROR] Failed to unmarshal. bgio: %#v err: %v", bgio, err)
		return nil, model.NewError(model.ErrOther, err.Error())
	}
	return colls, nil
}

// SelectOneCollection loads a single collection from the database
func (p Persister) SelectOneCollection(ctx context.Context, id uint32) (*model.Collection, *model.Error) {
	var coll model.Collection
	sid := strconv.FormatInt(int64(id), 10)
	qi := &dynamodb.QueryInput{
		TableName:              p.tableName,
		KeyConditionExpression: aws.String(pkName + "= :pk and begins_with(" + skName + ", :sk)"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":pk": {
				N: aws.String(sid),
			},
			":sk": {
				S: aws.String(collectionType),
			},
		},
	}
	qo, err := p.svc.Query(qi)
	if err != nil {
		log.Printf("[ERROR] Failed to get collection. qi: %#v err: %v", qi, err)
		return nil, model.NewError(model.ErrOther, err.Error())
	}
	// log.Printf("[DEBUG] qo = %#v", qo)
	found := false
	for _, item := range qo.Items {
		if *item[skName].S == collectionType {
			found = true
			err = dynamodbattribute.UnmarshalMap(item, &coll)
			if err != nil {
				log.Printf("[ERROR] Failed to unmarshal collections. qo: %#v err: %v", qo, err)
				return nil, model.NewError(model.ErrOther, err.Error())
			}
		}
	}
	if !found {
		return nil, model.NewError(model.ErrNotFound, sid)
	}
	// The category IDs are stored in the main record, so we could do just a GetItem. They're
	// stored redundantly so that we can query them. By using them here, we can test to make sure
	// they're working correctly, but to avoid getting two copies of each, we null out the slice here.
	itemCategories := coll.Categories
	coll.Categories = nil
	for _, item := range qo.Items {
		if strings.HasPrefix(*item[skName].S, collectionCategoryPrefix) {
			id := strings.TrimPrefix(*item[skName].S, collectionCategoryPrefix)
			categoryID, err := strconv.ParseUint(id, 10, 32)
			if err != nil {
				log.Printf("[ERROR] Failed to unmarshal category ID %s: %v", id, err)
				return nil, model.NewError(model.ErrOther, err.Error())
			}
			// log.Printf("[DEBUG] Category ID %s", id)
			coll.Categories = append(coll.Categories, uint32(categoryID))
		}
	}
	// Compare the category slices
	if !compareIDs(itemCategories, coll.Categories) {
		return nil, model.NewError(model.ErrOther, fmt.Sprintf("Internal error: DynamoDB categories don't match for collection ID %d.\n %#v != %#v",
			coll.ID, itemCategories, coll.Categories))
	}
	// log.Printf("[DEBUG] Got collection %#v", coll)
	return &coll, nil
}

// compareIDs returns true if the two slices contain the same members, regardless of order.
func compareIDs(s1, s2 []uint32) bool {
	if len(s1) != len(s2) {
		return false
	}
	m1 := map[uint32]bool{}
	for _, id := range s1 {
		m1[id] = true
	}
	for _, id := range s2 {
		if !m1[id] {
			return false
		}
	}
	return true
}

// InsertCollection inserts a Collection into the database and returns the inserted Collection
func (p Persister) InsertCollection(ctx context.Context, in model.CollectionIn) (*model.Collection, *model.Error) {
	var coll model.Collection
	var err error
	coll.ID, err = p.GetSequenceValue()
	if err != nil {
		return nil, model.NewError(model.ErrOther, err.Error())
	}
	coll.Type = collectionType
	coll.CollectionIn = in
	now := time.Now().Truncate(0)
	coll.InsertTime = now
	coll.LastUpdateTime = now

	// Each category requires a condition check and a put
	itemLen := 2*len(in.Categories) + 1
	if itemLen > 25 {
		// TODO: Consider relaxing this somehow
		return nil, model.NewError(model.ErrOther, fmt.Sprintf("Unable to insert a collection with more than 12 categories (%#v)", in))
	}

	avs, err := dynamodbattribute.MarshalMap(coll)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal collection %#v: %v", coll, err)
		return nil, model.NewError(model.ErrOther, err.Error())
	}

	twi := make([]*dynamodb.TransactWriteItem, itemLen)
	item := 0
	// log.Printf("[DEBUG] Creating category conditions")
	for _, catID := range in.Categories {
		twi[item] = &dynamodb.TransactWriteItem{
			ConditionCheck: &dynamodb.ConditionCheck{
				TableName: p.tableName,
				Key: map[string]*dynamodb.AttributeValue{
					pkName: {
						N: aws.String(strconv.FormatInt(int64(catID), 10)),
					},
					skName: {
						S: aws.String(categoryType),
					},
				},
				ConditionExpression: aws.String("attribute_exists(" + pkName + ") AND attribute_exists(" + skName + ")"),
			},
		}
		// log.Printf("[DEBUG] twi[%d] = %#v", item, twi[item])
		item++
	}
	// log.Printf("[DEBUG] Creating collection")
	twi[item] = &dynamodb.TransactWriteItem{
		Put: &dynamodb.Put{
			TableName:           p.tableName,
			Item:                avs,
			ConditionExpression: aws.String("attribute_not_exists(" + pkName + ")"), // Make duplicate insert fail
		},
	}
	// log.Printf("[DEBUG] twi[%d] = %#v", item, twi[item])
	item++
	// log.Printf("[DEBUG] Creating collection_categories")
	for _, catID := range in.Categories {
		twi[item] = &dynamodb.TransactWriteItem{
			Put: &dynamodb.Put{
				TableName: p.tableName,
				Item: map[string]*dynamodb.AttributeValue{
					pkName: {N: aws.String(strconv.FormatInt(int64(coll.ID), 10))},
					skName: {S: aws.String(collectionCategoryPrefix + strconv.FormatInt(int64(catID), 10))},
				},
				ConditionExpression: aws.String("attribute_not_exists(" + pkName + ") AND attribute_not_exists(" + skName + ")"), // Make duplicate insert fail
			},
		}
		// log.Printf("[DEBUG] twi[%d] = %#v", item, twi[item])
		item++
	}
	twii := &dynamodb.TransactWriteItemsInput{
		TransactItems: twi,
	}
	// log.Printf("[DEBUG] Executing TransactWriteItems(%#v)", twii)
	_, err = p.svc.TransactWriteItems(twii)
	if err != nil {
		switch e := err.(type) {
		case *dynamodb.TransactionCanceledException:
			for i, r := range e.CancellationReasons {
				if r != nil && *r.Code == "ConditionalCheckFailed" {
					switch {
					case i < len(in.Categories):
						// Theese items are the Category ID reference checks
						return nil, model.NewError(model.ErrBadReference, strconv.FormatInt(int64(in.Categories[i]), 10), categoryType)
					case i == len(in.Categories):
						return nil, model.NewError(model.ErrOther, fmt.Sprintf("Insert failed. Collection ID %d already exists", coll.ID))
					default: // i > len(in.Categories)
						// These are the collection_category puts.
						log.Printf("[ERROR] Failed to put collection %#v. twii: %#v err: %v", coll, twii, err)
						return nil, model.NewError(model.ErrOther, err.Error())
					}
				}
			}
		default:
			log.Printf("[ERROR] Failed to put collection %#v. twii: %#v err: %v", coll, twii, err)
			return nil, model.NewError(model.ErrOther, err.Error())
		}
	}
	return &coll, nil
}

// UpdateCollection updates a Collection in the database and returns the updated Collection
func (p Persister) UpdateCollection(ctx context.Context, id uint32, in model.Collection) (*model.Collection, *model.Error) {
	var coll model.Collection
	var err error
	coll = in
	coll.ID = id
	coll.Type = collectionType
	coll.LastUpdateTime = time.Now().Truncate(0)
	in.ID = id

	// Unfortunately we have to read before writing so that we know which categories have changed.
	// This isn't transactional so there's the possibility of a race and the two lists of categories
	// getting out of sync. The good news is that this operation is idempotent, so that it should
	// "self-heal", i.e. the next update should re-sync the lists.
	current, e := p.SelectOneCollection(ctx, id)
	if e != nil {
		return nil, e
	}
	currentCategories := map[uint32]bool{}
	for _, c := range current.Categories {
		currentCategories[c] = true
	}
	toAdd := []uint32{}
	inCategories := map[uint32]bool{}
	for _, c := range in.Categories {
		inCategories[c] = true
		if !currentCategories[c] {
			toAdd = append(toAdd, c)
		}
	}
	toRemove := []uint32{}
	for _, c := range current.Categories {
		if !inCategories[c] {
			toRemove = append(toRemove, c)
		}
	}
	// log.Printf("[DEBUG] Update categories to %#v, add %#v, to remove %#v", in.Categories, toAdd, toRemove)

	// Each new category reference requires a condition check and a put; each remove one requires a delete
	// Worst case here is that if you already have 12 categories (the insert limit), you can only replace them
	// with 6 new ones. (len(toAdd) = 6, len(toRemove) = 12)
	// Should we just go with a hard limit of 6?
	itemLen := 2*len(toAdd) + len(toRemove) + 1
	if itemLen > 25 {
		// TODO: Consider relaxing this somehow
		return nil, model.NewError(model.ErrOther, fmt.Sprintf("Unable to update collection; too many category changes (%#v)", in))
	}

	avs, err := dynamodbattribute.MarshalMap(coll)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal collection %#v: %v", coll, err)
		return nil, model.NewError(model.ErrOther, err.Error())
	}

	twi := make([]*dynamodb.TransactWriteItem, itemLen)
	item := 0
	// log.Printf("[DEBUG] Creating category conditions")
	for _, catID := range toAdd {
		twi[item] = &dynamodb.TransactWriteItem{
			ConditionCheck: &dynamodb.ConditionCheck{
				TableName: p.tableName,
				Key: map[string]*dynamodb.AttributeValue{
					pkName: {
						N: aws.String(strconv.FormatInt(int64(catID), 10)),
					},
					skName: {
						S: aws.String(categoryType),
					},
				},
				ConditionExpression: aws.String("attribute_exists(" + pkName + ") AND attribute_exists(" + skName + ")"),
			},
		}
		// log.Printf("[DEBUG] twi[%d] = %#v", item, twi[item])
		item++
	}
	lastUpdateTime := in.LastUpdateTime.Format(time.RFC3339Nano)
	twi[item] = &dynamodb.TransactWriteItem{
		Put: &dynamodb.Put{
			TableName:           p.tableName,
			Item:                avs,
			ConditionExpression: aws.String("last_update_time = :lastUpdateTime"), // Only allow updates, and use optimistic concurrency
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":lastUpdateTime": {S: &lastUpdateTime},
			},
		},
	}
	item++
	for _, catID := range toRemove {
		twi[item] = &dynamodb.TransactWriteItem{
			Delete: &dynamodb.Delete{
				TableName: p.tableName,
				Key: map[string]*dynamodb.AttributeValue{
					pkName: {N: aws.String(strconv.FormatInt(int64(coll.ID), 10))},
					skName: {S: aws.String(collectionCategoryPrefix + strconv.FormatInt(int64(catID), 10))},
				},
			},
		}
		// log.Printf("[DEBUG] twi[%d] = %#v", item, twi[item])
		item++
	}
	// log.Printf("[DEBUG] Creating collection_categories")
	for _, catID := range toAdd {
		twi[item] = &dynamodb.TransactWriteItem{
			Put: &dynamodb.Put{
				TableName: p.tableName,
				Item: map[string]*dynamodb.AttributeValue{
					pkName: {N: aws.String(strconv.FormatInt(int64(coll.ID), 10))},
					skName: {S: aws.String(collectionCategoryPrefix + strconv.FormatInt(int64(catID), 10))},
				},
			},
		}
		// log.Printf("[DEBUG] twi[%d] = %#v", item, twi[item])
		item++
	}
	twii := &dynamodb.TransactWriteItemsInput{
		TransactItems: twi,
	}
	// log.Printf("[DEBUG] Executing TransactWriteItems(%#v)", twii)
	_, err = p.svc.TransactWriteItems(twii)
	if err != nil {
		switch e := err.(type) {
		case *dynamodb.TransactionCanceledException:
			for i, r := range e.CancellationReasons {
				if r != nil && *r.Code == "ConditionalCheckFailed" {
					switch {
					case i < len(toAdd):
						// These are the category condition checks
						return nil, model.NewError(model.ErrBadReference, strconv.FormatInt(int64(toAdd[i]), 10), categoryType)
					case i == len(toAdd):
						// This is the actual put, so an error here is due to lastUpdateTime not matching
						return nil, model.NewError(model.ErrConcurrentUpdate, current.LastUpdateTime.Format(time.RFC3339Nano), lastUpdateTime)
					default: // i > len(toAdd)
						// These are the collection_category updates
						log.Printf("[ERROR] Failed to put collection %#v. twii: %#v err: %v", coll, twii, err)
						return nil, model.NewError(model.ErrOther, err.Error())
					}
				}
			}
		default:
			log.Printf("[ERROR] Failed to put collection %#v. twii: %#v err: %v", coll, twii, err)
			return nil, model.NewError(model.ErrOther, err.Error())
		}
	}
	return &coll, nil //model.NewError(model.ErrOther, err.Error())
}

// DeleteCollection deletes a Collection
func (p Persister) DeleteCollection(ctx context.Context, id uint32) *model.Error {
	// Read to get the category list, so that we can delete them
	current, e := p.SelectOneCollection(ctx, id)
	if e != nil {
		if e.Code == model.ErrNotFound {
			return nil
		}
		return e
	}
	twi := make([]*dynamodb.TransactWriteItem, len(current.Categories)+1)
	item := 0
	for _, catID := range current.Categories {
		twi[item] = &dynamodb.TransactWriteItem{
			Delete: &dynamodb.Delete{
				TableName: p.tableName,
				Key: map[string]*dynamodb.AttributeValue{
					pkName: {N: aws.String(strconv.FormatInt(int64(id), 10))},
					skName: {S: aws.String(collectionCategoryPrefix + strconv.FormatInt(int64(catID), 10))},
				},
			},
		}
		item++
	}
	twi[item] = &dynamodb.TransactWriteItem{
		Delete: &dynamodb.Delete{
			TableName: p.tableName,
			Key: map[string]*dynamodb.AttributeValue{
				pkName: {N: aws.String(strconv.FormatInt(int64(id), 10))},
				skName: {S: aws.String(collectionType)},
			},
		},
	}
	item++
	twii := &dynamodb.TransactWriteItemsInput{
		TransactItems: twi,
	}
	// log.Printf("[DEBUG] Executing TransactWriteItems(%#v)", twii)
	_, err := p.svc.TransactWriteItems(twii)
	if err != nil {
		return model.NewError(model.ErrOther, err.Error())
	}
	return nil
}
