package dynamo

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

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

// getSequenceValue returns a unique sequence value
func (p *Persister) getSequenceValue() (uint32, error) {
	uii := &dynamodb.UpdateItemInput{
		TableName: p.tableName,
		Key: map[string]*dynamodb.AttributeValue{
			"pk": {N: aws.String("1")},
			"sk": {S: aws.String("sequence")},
		},
		UpdateExpression: aws.String("ADD sequenceValue :i"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":i": {N: aws.String("1")},
		},
		ReturnValues: aws.String("ALL_NEW"),
	}
	uio, err := p.svc.UpdateItem(uii)
	if err != nil {
		return 0, err
	}
	i, err := strconv.ParseUint(*uio.Attributes["sequenceValue"].N, 10, 32)
	return uint32(i), err
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
				AttributeName: aws.String("pk"),
				AttributeType: aws.String("N"),
			},
			{
				AttributeName: aws.String("sk"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("data"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("pk"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("sk"),
				KeyType:       aws.String("RANGE"),
			},
		},
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			{
				IndexName: aws.String("sk_data"),
				KeySchema: []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("sk"),
						KeyType:       aws.String("HASH"),
					},
					{
						AttributeName: aws.String("data"),
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

func translateError(err error) error {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		// case dynamodb.ErrCodeProvisionedThroughputExceededException:
		// 	fmt.Println(dynamodb.ErrCodeProvisionedThroughputExceededException, aerr.Error())
		// case dynamodb.ErrCodeResourceNotFoundException:
		// 	fmt.Println(dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
		// case dynamodb.ErrCodeRequestLimitExceeded:
		// 	fmt.Println(dynamodb.ErrCodeRequestLimitExceeded, aerr.Error())
		// case dynamodb.ErrCodeInternalServerError:
		// 	fmt.Println(dynamodb.ErrCodeInternalServerError, aerr.Error())
		default:
			log.Printf("[INFO] Untranslated DynamoDB error: %v", aerr)
		}
	} else {
		log.Printf("[INFO] Untranslated DynamoDB error: %v", err)
	}
	return err
}
