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
	"github.com/ourrootsorg/cms-server/persist"
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
		return posts, translateError(err)
	}
	err = dynamodbattribute.UnmarshalListOfMaps(qo.Items, &posts)
	if err != nil {
		log.Printf("[ERROR] Failed to unmarshal posts. qo: %#v err: %v", qo, err)
		return posts, translateError(err)
	}
	return posts, nil
	// rows, err := p.db.QueryContext(ctx, "SELECT id, post_id, body, insert_time, last_update_time FROM post")
	// if err != nil {
	// 	return nil, translateError(err)
	// }
	// defer rows.Close()
	// posts := make([]model.Post, 0)
	// for rows.Next() {
	// 	var post model.Post
	// 	err := rows.Scan(&post.ID, &post.Post, &post.PostBody, &post.InsertTime, &post.LastUpdateTime)
	// 	if err != nil {
	// 		return nil, translateError(err)
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
				N: aws.String(strconv.FormatInt(int64(id), 10)),
			},
			skName: {
				S: aws.String(postType),
			},
		},
	}
	gio, err := p.svc.GetItem(gii)
	if err != nil {
		log.Printf("[ERROR] Failed to get post. qi: %#v err: %v", gio, err)
		return nil, translateError(err)
	}
	if gio.Item == nil {
		return nil, model.NewError(model.ErrNotFound, strconv.Itoa(int(id)))
	}
	err = dynamodbattribute.UnmarshalMap(gio.Item, &post)
	if err != nil {
		log.Printf("[ERROR] Failed to unmarshal. qo: %#v err: %v", gio, err)
		return nil, translateError(err)
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
	// 	return nil, translateError(err)
	// }
}

// InsertPost inserts a PostBody into the database and returns the inserted Post
func (p Persister) InsertPost(ctx context.Context, in model.PostIn) (*model.Post, error) {
	var post model.Post
	var err error
	post.ID, err = p.GetSequenceValue()
	if err != nil {
		return nil, translateError(err)
	}
	post.Type = postType
	post.PostIn = in
	now := time.Now().Truncate(0)
	post.InsertTime = now
	post.LastUpdateTime = now

	avs, err := dynamodbattribute.MarshalMap(post)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal post %#v: %v", post, err)
		return nil, translateError(err)
	}

	twi := make([]*dynamodb.TransactWriteItem, 2)
	twi[0] = &dynamodb.TransactWriteItem{
		ConditionCheck: &dynamodb.ConditionCheck{
			TableName: p.tableName,
			Key: map[string]*dynamodb.AttributeValue{
				pkName: {
					N: aws.String(strconv.FormatInt(int64(post.Collection), 10)),
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
				if r != nil && *r.Code == "ConditionalCheckFailed" {
					switch i {
					case 0:
						// This is the Collection ID reference check
						return nil, persist.ErrForeignKeyViolation
					case 1:
						return nil, model.NewError(model.ErrOther, fmt.Sprintf("Insert failed. Post ID %d already exists", post.ID))
					default: // i >= 2
						// Should never happen
						log.Printf("[ERROR] Failed to put post %#v. twii: %#v err: %v", post, twii, err)
						return nil, translateError(err)
					}
				}
			}
		default:
			log.Printf("[ERROR] Failed to put post %#v. twii: %#v err: %v", post, twii, err)
			return nil, translateError(err)
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
		return nil, translateError(err)
	}

	twi := make([]*dynamodb.TransactWriteItem, 2)
	twi[0] = &dynamodb.TransactWriteItem{
		ConditionCheck: &dynamodb.ConditionCheck{
			TableName: p.tableName,
			Key: map[string]*dynamodb.AttributeValue{
				pkName: {
					N: aws.String(strconv.FormatInt(int64(post.Collection), 10)),
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
				if r != nil && *r.Code == "ConditionalCheckFailed" {
					switch i {
					case 0:
						// This is the Collection ID reference check
						return nil, persist.ErrForeignKeyViolation
					case 1:
						// This is the actual put, so an error here is due to either lastUpdateTime not matching or the item not existing
						// Do a select to distinguish the cases
						current, err := p.SelectOnePost(ctx, id)
						if err != nil {
							if err == persist.ErrNoRows {
								return nil, model.NewError(model.ErrNotFound, strconv.Itoa(int(id)))
							}
							return nil, translateError(err)
						}
						return nil, model.NewError(model.ErrConcurrentUpdate, current.LastUpdateTime.Format(time.RFC3339Nano), lastUpdateTime)
					default: // i >= 2
						// Should never happen
						log.Printf("[ERROR] Failed to put post %#v. twii: %#v err: %v", post, twii, err)
						return nil, translateError(err)
					}
				}
			}
		default:
			log.Printf("[ERROR] Failed to put post %#v. twii: %#v err: %v", post, twii, err)
			return nil, translateError(err)
		}
	}
	return &post, nil
	// err := p.db.QueryRowContext(ctx,
	// 	`UPDATE post SET body = $1, post_id = $2, last_update_time = CURRENT_TIMESTAMP
	// 	 WHERE id = $3 AND last_update_time = $4
	// 	 RETURNING id, post_id, body, insert_time, last_update_time`,
	// 	in.PostBody, in.Post, id, in.LastUpdateTime).
	// 	Scan(
	// 		&post.ID,
	// 		&post.Post,
	// 		&post.PostBody,
	// 		&post.InsertTime,
	// 		&post.LastUpdateTime,
	// 	)
	// if err != nil && err == sql.ErrNoRows {
	// 	// Either non-existent or last_update_time didn't match
	// 	c, _ := p.SelectOnePost(ctx, id)
	// 	if c.ID == id {
	// 		// Row exists, so it must be a non-matching update time
	// 		return nil, model.NewError(model.ErrConcurrentUpdate, c.LastUpdateTime.String(), in.LastUpdateTime.String())
	// 	}
	// 	return nil, model.NewError(model.ErrNotFound, strconv.Itoa(int(id)))
	// }
}

// DeletePost deletes a Post
func (p Persister) DeletePost(ctx context.Context, id uint32) error {
	dii := &dynamodb.DeleteItemInput{
		TableName: p.tableName,
		Key: map[string]*dynamodb.AttributeValue{
			pkName: {N: aws.String(strconv.FormatInt(int64(id), 10))},
			skName: {S: aws.String(collectionType)},
		},
	}
	_, err := p.svc.DeleteItem(dii)
	if err != nil {
		return translateError(err)
	}
	// _, err := p.db.ExecContext(ctx, "DELETE FROM post WHERE id = $1", id)
	// return translateError(err)
	return nil
}
