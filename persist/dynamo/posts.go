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

const postType = "post"

// SelectPosts selects all posts
func (p Persister) SelectPosts(ctx context.Context) ([]model.Post, error) {
	qi := &dynamodb.QueryInput{
		TableName:              p.tableName,
		IndexName:              aws.String(gsiName),
		KeyConditionExpression: aws.String(skName + " = :sk"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":sk": {
				S: aws.String(postType),
			},
		},
	}
	posts := make([]model.Post, 0)
	qo, err := p.svc.Query(qi)
	if err != nil {
		log.Printf("[ERROR] Failed to get posts. qi: %#v err: %v", qi, err)
		return posts, model.NewError(model.ErrOther, err.Error())
	}
	err = dynamodbattribute.UnmarshalListOfMaps(qo.Items, &posts)
	if err != nil {
		log.Printf("[ERROR] Failed to unmarshal posts. qo: %#v err: %v", qo, err)
		return posts, model.NewError(model.ErrOther, err.Error())
	}
	return posts, nil
	// rows, err := p.db.QueryContext(ctx, "SELECT id, post_id, body, insert_time, last_update_time FROM post")
	// if err != nil {
	// 	return nil, model.NewError(model.ErrOther, err.Error())
	// }
	// defer rows.Close()
	// posts := make([]model.Post, 0)
	// for rows.Next() {
	// 	var post model.Post
	// 	err := rows.Scan(&post.ID, &post.Post, &post.PostBody, &post.InsertTime, &post.LastUpdateTime)
	// 	if err != nil {
	// 		return nil, model.NewError(model.ErrOther, err.Error())
	// 	}
	// 	posts = append(posts, post)
	// }
}

// SelectOnePost selects a single post
func (p Persister) SelectOnePost(ctx context.Context, id uint32) (*model.Post, error) {
	var post model.Post
	gii := &dynamodb.GetItemInput{
		TableName: p.tableName,
		Key: map[string]*dynamodb.AttributeValue{
			pkName: {
				S: aws.String(strconv.FormatInt(int64(id), 10)),
			},
			skName: {
				S: aws.String(postType),
			},
		},
	}
	gio, err := p.svc.GetItem(gii)
	if err != nil {
		log.Printf("[ERROR] Failed to get post. qi: %#v err: %v", gio, err)
		return nil, model.NewError(model.ErrOther, err.Error())
	}
	if gio.Item == nil {
		return nil, model.NewError(model.ErrNotFound, strconv.Itoa(int(id)))
	}
	err = dynamodbattribute.UnmarshalMap(gio.Item, &post)
	if err != nil {
		log.Printf("[ERROR] Failed to unmarshal. qo: %#v err: %v", gio, err)
		return nil, model.NewError(model.ErrOther, err.Error())
	}
	return &post, nil
	// err := p.db.QueryRowContext(ctx, "SELECT id, post_id, body, insert_time, last_update_time FROM post WHERE id=$1", id).Scan(
	// 	&post.ID,
	// 	&post.Post,
	// 	&post.PostBody,
	// 	&post.InsertTime,
	// 	&post.LastUpdateTime,
	// )
	// if err != nil {
	// 	return nil, model.NewError(model.ErrOther, err.Error())
	// }
}

// InsertPost inserts a PostBody into the database and returns the inserted Post
func (p Persister) InsertPost(ctx context.Context, in model.PostIn) (*model.Post, error) {
	var post model.Post
	var err error
	post.ID, err = p.GetSequenceValue()
	if err != nil {
		return nil, model.NewError(model.ErrOther, err.Error())
	}
	post.Type = postType
	post.PostIn = in
	now := time.Now().Truncate(0)
	post.InsertTime = now
	post.LastUpdateTime = now

	avs, err := dynamodbattribute.MarshalMap(post)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal post %#v: %v", post, err)
		return nil, model.NewError(model.ErrOther, err.Error())
	}

	twi := make([]*dynamodb.TransactWriteItem, 2)
	twi[0] = &dynamodb.TransactWriteItem{
		ConditionCheck: &dynamodb.ConditionCheck{
			TableName: p.tableName,
			Key: map[string]*dynamodb.AttributeValue{
				pkName: {
					S: aws.String(strconv.FormatInt(int64(post.Collection), 10)),
				},
				skName: {
					S: aws.String(collectionType),
				},
			},
			ConditionExpression: aws.String("attribute_exists(" + pkName + ") AND attribute_exists(" + skName + ")"),
		},
	}
	// log.Printf("[DEBUG] Creating post")
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
							// This is the Collection ID reference check
							return nil, model.NewError(model.ErrBadReference, strconv.FormatInt(int64(post.Collection), 10), collectionType)
						case 1:
							return nil, model.NewError(model.ErrOther, fmt.Sprintf("Insert failed. Post ID %d already exists", post.ID))
						default: // i >= 2
							// Should never happen
							log.Printf("[ERROR] Failed to put post %#v. twii: %#v err: %v", post, twii, err)
							return nil, model.NewError(model.ErrOther, err.Error())
						}
					} else if *r.Code == "TransactionConflict" {
						log.Printf("[ERROR] TransactionConflict when putting post %#v. twii: %#v err: %v", post, twii, err)
						return nil, model.NewError(model.ErrConflict)
					} else {
						log.Printf("[ERROR] Failed to put post %#v. twii: %#v err: %v", post, twii, err)
						return nil, model.NewError(model.ErrOther, err.Error())
					}
				}
			}
		default:
			log.Printf("[ERROR] Failed to put post %#v. twii: %#v err: %v", post, twii, err)
			return nil, model.NewError(model.ErrOther, err.Error())
		}
	}
	return &post, nil
	// err := p.db.QueryRowContext(ctx,
	// 	`INSERT INTO post (post_id, body)
	// 	 VALUES ($1, $2)
	// 	 RETURNING id, post_id, body, insert_time, last_update_time`,
	// 	in.Post, in.PostBody).
	// 	Scan(
	// 		&post.ID,
	// 		&post.Post,
	// 		&post.PostBody,
	// 		&post.InsertTime,
	// 		&post.LastUpdateTime,
	// 	)
}

// UpdatePost updates a Post in the database and returns the updated Post
func (p Persister) UpdatePost(ctx context.Context, id uint32, in model.Post) (*model.Post, error) {
	var post model.Post
	var err error
	post = in
	post.Type = postType
	post.LastUpdateTime = time.Now().Truncate(0)

	avs, err := dynamodbattribute.MarshalMap(post)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal post %#v: %v", post, err)
		return nil, model.NewError(model.ErrOther, err.Error())
	}

	twi := make([]*dynamodb.TransactWriteItem, 2)
	twi[0] = &dynamodb.TransactWriteItem{
		ConditionCheck: &dynamodb.ConditionCheck{
			TableName: p.tableName,
			Key: map[string]*dynamodb.AttributeValue{
				pkName: {
					S: aws.String(strconv.FormatInt(int64(post.Collection), 10)),
				},
				skName: {
					S: aws.String(collectionType),
				},
			},
			ConditionExpression: aws.String("attribute_exists(" + pkName + ") AND attribute_exists(" + skName + ")"),
		},
	}
	// log.Printf("[DEBUG] Creating post")
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
		switch e := err.(type) {
		case *dynamodb.TransactionCanceledException:
			for i, r := range e.CancellationReasons {
				if r != nil && *r.Code != "None" {
					if *r.Code == "ConditionalCheckFailed" {
						switch i {
						case 0:
							// This is the Collection ID reference check
							return nil, model.NewError(model.ErrBadReference, strconv.FormatInt(int64(post.Collection), 10), collectionType)
						case 1:
							// This is the actual put, so an error here is due to either lastUpdateTime not matching or the item not existing
							// Do a select to distinguish the cases
							current, e := p.SelectOnePost(ctx, id)
							if e != nil {
								return nil, e
							}
							return nil, model.NewError(model.ErrConcurrentUpdate, current.LastUpdateTime.Format(time.RFC3339Nano), lastUpdateTime)
						default: // i >= 2
							// Should never happen
							log.Printf("[ERROR] Failed to put post %#v. twii: %#v err: %v", post, twii, err)
							return nil, model.NewError(model.ErrOther, err.Error())
						}
					} else if *r.Code == "TransactionConflict" {
						log.Printf("[ERROR] TransactionConflict when putting post %#v. twii: %#v err: %v", post, twii, err)
						return nil, model.NewError(model.ErrConflict)
					} else {
						log.Printf("[ERROR] Failed to put post %#v. twii: %#v err: %v", post, twii, err)
						return nil, model.NewError(model.ErrOther, err.Error())
					}
				}
			}
		default:
			log.Printf("[ERROR] Failed to put post %#v. twii: %#v err: %v", post, twii, err)
			return nil, model.NewError(model.ErrOther, err.Error())
		}
	}
	return &post, nil
}

// DeletePost deletes a Post
func (p Persister) DeletePost(ctx context.Context, id uint32) error {
	dii := &dynamodb.DeleteItemInput{
		TableName: p.tableName,
		Key: map[string]*dynamodb.AttributeValue{
			pkName: {S: aws.String(strconv.FormatInt(int64(id), 10))},
			skName: {S: aws.String(postType)},
		},
	}
	_, err := p.svc.DeleteItem(dii)
	if err != nil {
		return model.NewError(model.ErrOther, err.Error())
	}
	return nil
}
