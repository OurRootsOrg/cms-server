package dynamo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/ourrootsorg/cms-server/model"
)

const userType = "user"

// RetrieveUser either retrieves a user record from the database, or creates the record if it doesn't
// already exist.
func (p Persister) RetrieveUser(ctx context.Context, in model.UserIn) (*model.User, error) {
	var user model.User
	var err error
	log.Printf("[DEBUG] Looking up subject '%s' in database", in.Subject)
	// Issuer and Subject are arbitrary strings, so there's no safe separator character to use to split them.
	// So we URL-encode them in the database.
	key := url.Values{}
	key.Add("iss", in.Issuer)
	key.Add("sub", in.Subject)

	qi := &dynamodb.QueryInput{
		TableName:              p.tableName,
		IndexName:              aws.String(gsiName),
		KeyConditionExpression: aws.String(skName + " = :sk and " + gsiSkName + " = :altSort"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":sk": {
				S: aws.String(userType),
			},
			":altSort": {
				S: aws.String(key.Encode()),
			},
		},
	}
	qo, err := p.svc.Query(qi)
	if err != nil {
		log.Printf("[DEBUG] Error looking up subject '%s' in database", in.Subject)
		return nil, translateError(err)
	}
	if *qo.Count == 1 {
		log.Printf("[DEBUG] Found subject '%s' in database", in.Subject)
		var users []model.User
		err = dynamodbattribute.UnmarshalListOfMaps(qo.Items, &users)
		if err != nil {
			log.Printf("[ERROR] Failed to unmarshal. qo: %#v err: %v", qo, err)
			return nil, translateError(err)
		}
		user = users[0]
		if !user.Enabled {
			msg := fmt.Sprintf("User '%d' is not enabled", user.ID)
			log.Printf("[DEBUG] %s", msg)
			return nil, errors.New(msg)
		}
		// We got a user
		log.Printf("[DEBUG] Returning enabled user '%#v'", user)
		return &user, nil
	}
	log.Printf("[DEBUG] No user with subject '%s' found in database, so creating one", in.Subject)

	user.ID, err = p.getSequenceValue()
	if err != nil {
		return nil, translateError(err)
	}
	user.Type = userType
	user.SortKey = key.Encode()
	user.UserBody = in.UserBody
	now := time.Now().Truncate(0)
	user.InsertTime = now
	user.LastUpdateTime = now

	avs, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal category %#v: %v", user, err)
		return nil, translateError(err)
	}

	pii := &dynamodb.PutItemInput{
		TableName:           p.tableName,
		Item:                avs,
		ConditionExpression: aws.String("attribute_not_exists(id)"), // Make duplicate insert fail
	}
	pio, err := p.svc.PutItem(pii)
	if err != nil {
		if compareToAWSError(err, dynamodb.ErrCodeConditionalCheckFailedException) {
			return nil, model.NewError(model.ErrOther, fmt.Sprintf("Insert failed. User ID %d already exists", user.ID))
		}
		log.Printf("[ERROR] Failed to put user %#v. pii: %#v err: %v", user, pii, err)
		return nil, translateError(err)
	}
	err = dynamodbattribute.UnmarshalMap(pio.Attributes, &user)
	if err != nil {
		return nil, translateError(err)
	}

	// err = p.db.QueryRowContext(ctx,
	// 	`INSERT INTO cms_user (body)
	// 	 VALUES ($1)
	// 	 RETURNING id, body, insert_time, last_update_time`,
	// 	in.UserBody).
	// 	Scan(
	// 		&user.ID,
	// 		&user.UserBody,
	// 		&user.InsertTime,
	// 		&user.LastUpdateTime,
	// 	)
	log.Printf("[DEBUG] Created user '%d'", user.ID)
	return &user, translateError(err)
}
