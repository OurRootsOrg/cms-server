package dynamo

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

const (
	// You can't use constants in struct tags (see https://github.com/golang/go/issues/4740),
	// so if these are ever changed, they must also be replaced in all `dynamodbav` tags.
	pkName    = "pk"
	skName    = "sk"
	gsiSkName = "altSort"
	gsiName   = "gsi_" + skName + "_" + gsiSkName

	idSeparator = "#"
)

// Reserved non-sequential IDs
// const (
// 	sequenceID = -1
// 	settingsID = -2
// )

// Persister persists the model objects to DynammoDB
type Persister struct {
	svc       *dynamodb.DynamoDB
	tableName *string
}

// NewPersister constructs a new Persister
func NewPersister(session *session.Session, tableName string) (Persister, error) {
	svc := dynamodb.New(session)
	err := ensureTableExists(svc, tableName)
	p := Persister{
		svc:       svc,
		tableName: &tableName,
	}
	return p, err
}

// GetSequenceValue returns a unique sequence value
func (p *Persister) GetSequenceValue() (uint32, error) {
	v, err := p.GetMultipleSequenceValues(1)
	return v[0], err
}

// GetMultipleSequenceValues returns a slice of unique sequence values
func (p *Persister) GetMultipleSequenceValues(cnt int) ([]uint32, error) {
	if cnt < 1 || cnt > 10000 {
		return nil, errors.New("Must request between 1 and 10,000 sequence values")
	}
	uii := &dynamodb.UpdateItemInput{
		TableName: p.tableName,
		Key: map[string]*dynamodb.AttributeValue{
			pkName: {S: aws.String("sequence")},
			skName: {S: aws.String("sequence")},
		},
		UpdateExpression: aws.String("ADD sequenceValue :i"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":i": {N: aws.String(strconv.Itoa(cnt))},
		},
		ReturnValues: aws.String("ALL_NEW"),
	}
	uio, err := p.svc.UpdateItem(uii)
	if err != nil {
		return nil, err
	}
	v, err := strconv.ParseUint(*uio.Attributes["sequenceValue"].N, 10, 32)
	if err != nil {
		return nil, err
	}
	ret := make([]uint32, cnt)
	for i := uint32(0); i < uint32(cnt); i++ {
		ret[i] = uint32(v) - uint32(cnt) + i + 1
	}
	return ret, err
}

func ensureTableExists(svc *dynamodb.DynamoDB, tableName string) error {
	// See if the table exists already
	err := waitForTable(svc, tableName)
	if err == nil {
		// Perhaps we should compare the schema to what we expect
		// if so, see https://github.com/dollarshaveclub/dynamo-drift
		return nil
	}
	if aerr, ok := err.(awserr.Error); ok {
		if aerr.Code() != dynamodb.ErrCodeResourceNotFoundException {
			return fmt.Errorf("[ERROR] Unexpected error when checking DynamoDB table: %v", err)
		}
	} else {
		return fmt.Errorf("[ERROR] Unexpected error when checking DynamoDB table: %v", err)
	}
	log.Printf("[INFO] Creating table %s", tableName)
	// Couldn't describe it, so try to create it
	cti := &dynamodb.CreateTableInput{
		TableName: &tableName,
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String(pkName),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String(skName),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String(gsiSkName),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String(pkName),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String(skName),
				KeyType:       aws.String("RANGE"),
			},
		},
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			{
				IndexName: aws.String(gsiName),
				KeySchema: []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String(skName),
						KeyType:       aws.String("HASH"),
					},
					{
						AttributeName: aws.String(gsiSkName),
						KeyType:       aws.String("RANGE"),
					},
				},
				Projection: &dynamodb.Projection{
					ProjectionType: aws.String("ALL"),
				},
				// TODO: How should these values be set/modified?
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
			},
		},
		// TODO: How should these values be set/modified?
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	}
	_, err = svc.CreateTable(cti)
	if err != nil {
		log.Printf("[ERROR] Failed to create table %s: %v", tableName, err)
		return err
	}
	return waitForTable(svc, tableName)
}
func waitForTable(svc *dynamodb.DynamoDB, tableName string) error {
	cnt := 0
	for {
		// See if the table exists
		dto, err := svc.DescribeTable(&dynamodb.DescribeTableInput{
			TableName: &tableName,
		})
		if err != nil {
			log.Printf("[INFO] Couldn't describe DynamoDB table %s: %v", tableName, err)
			return err
		}
		if *dto.Table.TableStatus == "ACTIVE" {
			log.Printf("[DEBUG] DynamoDB table %s is ACTIVE", tableName)
			return nil
		}
		if *dto.Table.TableStatus == "CREATING" || *dto.Table.TableStatus == "UPDATING" && cnt < 10 {
			log.Printf("[DEBUG] Waiting for DynamoDB table %s to be ACTIVE. Current status is %s", tableName, *dto.Table.TableStatus)
			// wait a bit before trying again
			time.Sleep(5 * time.Second)
		} else {
			return fmt.Errorf("Unexpected status while waiting for DynamoDB table %s to be ready: %s", tableName, *dto.Table.TableStatus)
		}
		cnt++
	}

}

func compareToAWSError(err error, awsErrorCode string) bool {
	if aerr, ok := err.(awserr.Error); ok {
		return aerr.Code() == awsErrorCode
	}
	return false
}
