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

const recordType = "record"

// SelectRecords selects all records
func (p Persister) SelectRecords(ctx context.Context) ([]model.Record, error) {
	qi := &dynamodb.QueryInput{
		TableName:              p.tableName,
		IndexName:              aws.String(gsiName),
		KeyConditionExpression: aws.String(skName + " = :sk"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":sk": {
				S: aws.String(recordType),
			},
		},
	}
	records := make([]model.Record, 0)
	qo, err := p.svc.Query(qi)
	if err != nil {
		log.Printf("[ERROR] Failed to get records. qi: %#v err: %v", qi, err)
		return records, model.NewError(model.ErrOther, err.Error())
	}
	err = dynamodbattribute.UnmarshalListOfMaps(qo.Items, &records)
	if err != nil {
		log.Printf("[ERROR] Failed to unmarshal records. qo: %#v err: %v", qo, err)
		return records, model.NewError(model.ErrOther, err.Error())
	}
	return records, nil
}

// SelectRecordsByID selects many records from a slice of IDs
func (p Persister) SelectRecordsByID(ctx context.Context, ids []uint32) ([]model.Record, error) {
	records := make([]model.Record, 0)
	if len(ids) == 0 {
		return records, nil
	}
	keys := make([]map[string]*dynamodb.AttributeValue, len(ids))
	for i, id := range ids {
		keys[i] = map[string]*dynamodb.AttributeValue{
			pkName: {
				S: aws.String(strconv.FormatInt(int64(id), 10)),
			},
			skName: {
				S: aws.String(recordType),
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
		log.Printf("[ERROR] Failed to get records. bgii: %#v err: %v", bgii, err)
		return nil, model.NewError(model.ErrOther, err.Error())
	}
	err = dynamodbattribute.UnmarshalListOfMaps(bgio.Responses[*p.tableName], &records)
	if err != nil {
		log.Printf("[ERROR] Failed to unmarshal. bgio: %#v err: %v", bgio, err)
		return nil, model.NewError(model.ErrOther, err.Error())
	}
	return records, nil
}

// SelectRecordsForPost selects all records for a Post
// This is not currently part of the persist interface, but it's here when we need it
func (p Persister) SelectRecordsForPost(ctx context.Context, postID uint32) ([]model.Record, error) {
	qi := &dynamodb.QueryInput{
		TableName:              p.tableName,
		IndexName:              aws.String(gsiName),
		KeyConditionExpression: aws.String(skName + " = :sk and " + gsiSkName + " = :gsiSk"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":sk": {
				S: aws.String(recordType),
			},
			":gsiSk": {
				S: aws.String(strconv.FormatInt(int64(postID), 10)),
			},
		},
	}
	records := make([]model.Record, 0)
	qo, err := p.svc.Query(qi)
	if err != nil {
		log.Printf("[ERROR] Failed to get records. qi: %#v err: %v", qi, err)
		return records, model.NewError(model.ErrOther, err.Error())
	}
	err = dynamodbattribute.UnmarshalListOfMaps(qo.Items, &records)
	if err != nil {
		log.Printf("[ERROR] Failed to unmarshal records. qo: %#v err: %v", qo, err)
		return records, model.NewError(model.ErrOther, err.Error())
	}
	return records, nil
}

// SelectOneRecord selects a single record by ID
func (p Persister) SelectOneRecord(ctx context.Context, id uint32) (*model.Record, error) {
	var record model.Record
	gii := &dynamodb.GetItemInput{
		TableName: p.tableName,
		Key: map[string]*dynamodb.AttributeValue{
			pkName: {
				S: aws.String(strconv.FormatInt(int64(id), 10)),
			},
			skName: {
				S: aws.String(recordType),
			},
		},
	}
	gio, err := p.svc.GetItem(gii)
	if err != nil {
		log.Printf("[ERROR] Failed to get record. qi: %#v err: %v", gio, err)
		return nil, model.NewError(model.ErrOther, err.Error())
	}
	if gio.Item == nil {
		return nil, model.NewError(model.ErrNotFound, strconv.Itoa(int(id)))
	}
	err = dynamodbattribute.UnmarshalMap(gio.Item, &record)
	if err != nil {
		log.Printf("[ERROR] Failed to unmarshal. qo: %#v err: %v", gio, err)
		return nil, model.NewError(model.ErrOther, err.Error())
	}
	return &record, nil
	// err := p.db.QueryRowContext(ctx, "SELECT id, record_id, body, insert_time, last_update_time FROM record WHERE id=$1", id).Scan(
	// 	&record.ID,
	// 	&record.Record,
	// 	&record.RecordBody,
	// 	&record.InsertTime,
	// 	&record.LastUpdateTime,
	// )
	// if err != nil {
	// 	return nil, model.NewError(model.ErrOther, err.Error())
	// }
}

// InsertRecord inserts a RecordBody into the database and returns the inserted Record
func (p Persister) InsertRecord(ctx context.Context, in model.RecordIn) (*model.Record, error) {
	var record model.Record
	var err error
	record.ID, err = p.GetSequenceValue()
	if err != nil {
		return nil, model.NewError(model.ErrOther, err.Error())
	}
	record.Type = recordType
	record.RecordIn = in
	now := time.Now().Truncate(0)
	record.InsertTime = now
	record.LastUpdateTime = now

	avs, err := dynamodbattribute.MarshalMap(record)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal record %#v: %v", record, err)
		return nil, model.NewError(model.ErrOther, err.Error())
	}

	twi := make([]*dynamodb.TransactWriteItem, 2)
	twi[0] = &dynamodb.TransactWriteItem{
		ConditionCheck: &dynamodb.ConditionCheck{
			TableName: p.tableName,
			Key: map[string]*dynamodb.AttributeValue{
				pkName: {
					S: aws.String(strconv.FormatInt(int64(record.Post), 10)),
				},
				skName: {
					S: aws.String(postType),
				},
			},
			ConditionExpression: aws.String("attribute_exists(" + pkName + ") AND attribute_exists(" + skName + ")"),
		},
	}
	twi[1] = &dynamodb.TransactWriteItem{
		Put: &dynamodb.Put{
			TableName:           p.tableName,
			Item:                avs,
			ConditionExpression: aws.String("attribute_not_exists(" + pkName + ")"), // Make duplicate insert fail
		},
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
				if r != nil && *r.Code != "None" {
					if *r.Code == "ConditionalCheckFailed" {
						switch i {
						case 0:
							// This is the Post ID reference check
							return nil, model.NewError(model.ErrBadReference, strconv.FormatInt(int64(record.Post), 10), postType)
						case 1:
							return nil, model.NewError(model.ErrOther, fmt.Sprintf("Insert failed. Record ID %d already exists", record.ID))
						default: // i >= 2
							// Should never happen
							log.Printf("[ERROR] Failed to put record %#v. twii: %#v err: %v", record, twii, err)
							return nil, model.NewError(model.ErrOther, err.Error())
						}
					} else if *r.Code == "TransactionConflict" {
						log.Printf("[ERROR] TransactionConflict when putting record %#v. twii: %#v err: %v", record, twii, err)
						return nil, model.NewError(model.ErrConflict)
					} else {
						log.Printf("[ERROR] Failed to put record %#v. twii: %#v err: %v", record, twii, err)
						return nil, model.NewError(model.ErrOther, err.Error())
					}
				}
			}
		default:
			log.Printf("[ERROR] Failed to put record %#v. twii: %#v err: %v", record, twii, err)
			return nil, model.NewError(model.ErrOther, err.Error())
		}
	}
	return &record, nil
}

// UpdateRecord updates a Record in the database and returns the updated Record
func (p Persister) UpdateRecord(ctx context.Context, id uint32, in model.Record) (*model.Record, error) {
	var record model.Record
	var err error
	record = in
	record.ID = id
	record.Type = recordType
	record.LastUpdateTime = time.Now().Truncate(0)

	avs, err := dynamodbattribute.MarshalMap(record)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal record %#v: %v", record, err)
		return nil, model.NewError(model.ErrOther, err.Error())
	}

	twi := make([]*dynamodb.TransactWriteItem, 2)
	twi[0] = &dynamodb.TransactWriteItem{
		ConditionCheck: &dynamodb.ConditionCheck{
			TableName: p.tableName,
			Key: map[string]*dynamodb.AttributeValue{
				pkName: {
					S: aws.String(strconv.FormatInt(int64(record.Post), 10)),
				},
				skName: {
					S: aws.String(postType),
				},
			},
			ConditionExpression: aws.String("attribute_exists(" + pkName + ") AND attribute_exists(" + skName + ")"),
		},
	}
	lastUpdateTime := in.LastUpdateTime.Format(time.RFC3339Nano)
	twi[1] = &dynamodb.TransactWriteItem{
		Put: &dynamodb.Put{
			TableName:           p.tableName,
			Item:                avs,
			ConditionExpression: aws.String("last_update_time = :lastUpdateTime"), // Only allow updates, and use optimistic concurrency
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":lastUpdateTime": {S: &lastUpdateTime},
			},
		},
	}
	twii := &dynamodb.TransactWriteItemsInput{
		TransactItems: twi,
	}
	// log.Printf("[DEBUG] Executing TransactWriteItems(%#v)", twii)
	_, err = p.svc.TransactWriteItems(twii)
	if err != nil {
		// log.Printf("[DEBUG] TransactWriteItems() err: %#v", err)
		switch e := err.(type) {
		case *dynamodb.TransactionCanceledException:
			for i, r := range e.CancellationReasons {
				if r != nil && *r.Code != "None" {
					if *r.Code == "ConditionalCheckFailed" {
						switch i {
						case 0:
							// This is the Post ID reference check
							return nil, model.NewError(model.ErrBadReference, strconv.FormatInt(int64(record.Post), 10), postType)
						case 1:
							// This is the actual put, so an error here is due to either lastUpdateTime not matching or the item not existing
							// Do a select to distinguish the cases
							current, e := p.SelectOneRecord(ctx, id)
							if e != nil {
								return nil, e
							}
							return nil, model.NewError(model.ErrConcurrentUpdate, current.LastUpdateTime.Format(time.RFC3339Nano), lastUpdateTime)
						default: // i >= 2
							// Should never happen
							log.Printf("[ERROR] Failed to put record %#v. twii: %#v err: %v", record, twii, err)
							return nil, model.NewError(model.ErrOther, err.Error())
						}
					} else if *r.Code == "TransactionConflict" {
						log.Printf("[ERROR] TransactionConflict when putting record %#v. twii: %#v err: %v", record, twii, err)
						return nil, model.NewError(model.ErrConflict)
					} else {
						log.Printf("[ERROR] Failed to put record %#v. twii: %#v err: %v", record, twii, err)
						return nil, model.NewError(model.ErrOther, err.Error())
					}
				}
			}
		default:
			log.Printf("[ERROR] Failed to put record %#v. twii: %#v err: %v", record, twii, err)
			return nil, model.NewError(model.ErrOther, err.Error())
		}
	}
	return &record, nil
}

// DeleteRecord deletes a Record
func (p Persister) DeleteRecord(ctx context.Context, id uint32) error {
	dii := &dynamodb.DeleteItemInput{
		TableName: p.tableName,
		Key: map[string]*dynamodb.AttributeValue{
			pkName: {S: aws.String(strconv.FormatInt(int64(id), 10))},
			skName: {S: aws.String(recordType)},
		},
	}
	_, err := p.svc.DeleteItem(dii)
	if err != nil {
		return model.NewError(model.ErrOther, err.Error())
	}
	return nil
}

// DeleteRecordsForPost deletes the Records associated with a Post
func (p Persister) DeleteRecordsForPost(ctx context.Context, postID uint32) error {
	records, err := p.SelectRecordsForPost(ctx, postID)
	if err != nil {
		return err
	}
	batchSize := 25
	ris := map[string][]*dynamodb.WriteRequest{}
	for _, r := range records {
		ris[*p.tableName] = append(ris[*p.tableName],
			&dynamodb.WriteRequest{
				DeleteRequest: &dynamodb.DeleteRequest{
					Key: map[string]*dynamodb.AttributeValue{
						pkName: {S: aws.String(strconv.FormatInt(int64(r.ID), 10))},
						skName: {S: aws.String(recordType)},
					},
				},
			},
		)
		// log.Printf("[DEBUG] DeleteRecordsForPost ris[%s][%d]: %#v", *p.tableName, i%batchSize, ris[*p.tableName][i%batchSize])

		if len(ris[*p.tableName]) == batchSize {
			log.Printf("[DEBUG] DeleteRecordsForPost, batch ris: %#v", ris)
			// Delete a batch
			err := p.deleteRecordBatch(ris)
			if err != nil {
				return model.NewError(model.ErrOther, err.Error())
			}
			// Get ready for next batch
			ris[*p.tableName] = []*dynamodb.WriteRequest{}
		}
	}
	if len(ris[*p.tableName]) > 0 {
		log.Printf("[DEBUG] DeleteRecordsForPost, final ris: %#v", ris)
		// Delete final batch
		err := p.deleteRecordBatch(ris)
		if err != nil {
			return model.NewError(model.ErrOther, err.Error())
		}
	}
	return nil
}

// SelectRecordHouseholdsForPost selects all record households for a post
func (p Persister) SelectRecordHouseholdsForPost(ctx context.Context, postID uint32) ([]model.RecordHousehold, error) {
	// TODO implement
	return nil, fmt.Errorf("SelectRecordHouseholdsForPost not implemented")
}

// SelectOneRecordHousehold selects one record household
func (p Persister) SelectOneRecordHousehold(ctx context.Context, postID uint32, householdID string) (*model.RecordHousehold, error) {
	// TODO implement
	return nil, fmt.Errorf("SelectOneRecordHousehold not implemented")
}

// InsertRecordHousehold inserts a RecordHouseholdIn into the database and returns a RecordHousehold
func (p Persister) InsertRecordHousehold(ctx context.Context, in model.RecordHouseholdIn) (*model.RecordHousehold, error) {
	// TODO implement
	return nil, fmt.Errorf("InsertRecordHousehold not implemented")
}

// DeleteRecordHouseholdsForPost deletes the record households associated with a Post
func (p Persister) DeleteRecordHouseholdsForPost(ctx context.Context, postID uint32) error {
	// TODO implement
	return nil // return nil so posts tests still work
}

func (p Persister) deleteRecordBatch(ris map[string][]*dynamodb.WriteRequest) error {
	var bwio *dynamodb.BatchWriteItemOutput
	var err error
	bwii := &dynamodb.BatchWriteItemInput{
		RequestItems: ris,
	}
	for bwio == nil || (len(bwio.UnprocessedItems) > 0 && err == nil) {
		if bwio != nil && len(bwio.UnprocessedItems) > 0 {
			bwii.RequestItems = bwio.UnprocessedItems
		}
		bwio, err = p.svc.BatchWriteItem(bwii)
	}
	return err
}
