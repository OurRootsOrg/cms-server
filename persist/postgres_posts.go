package persist

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ourrootsorg/cms-server/model"
)

// SelectPosts selects all posts
func (p PostgresPersister) SelectPosts(ctx context.Context) ([]model.Post, error) {
	rows, err := p.db.QueryContext(ctx, "SELECT id, collection_id, body, insert_time, last_update_time FROM post")
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()
	posts := make([]model.Post, 0)
	for rows.Next() {
		var dbID int32
		var collectionID int32
		var post model.Post
		err := rows.Scan(&dbID, &collectionID, &post.PostBody, &post.InsertTime, &post.LastUpdateTime)
		if err != nil {
			return nil, translateError(err)
		}
		post.ID = model.MakePostID(dbID)
		post.Collection = model.MakeCollectionID(collectionID)
		posts = append(posts, post)
	}
	return posts, nil
}

// SelectOnePost selects a single post
func (p PostgresPersister) SelectOnePost(ctx context.Context, id string) (model.Post, error) {
	var post model.Post
	var dbID int32
	var catID int32
	n, err := fmt.Sscanf(id+"\n", p.pathPrefix+model.PostIDFormat+"\n", &dbID)
	if err != nil || n != 1 {
		// Bad ID
		return post, model.NewError(model.ErrNotFound, id)
	}
	err = p.db.QueryRowContext(ctx, "SELECT id, collection_id, body, insert_time, last_update_time FROM post WHERE id=$1", dbID).Scan(
		&dbID,
		&catID,
		&post.PostBody,
		&post.InsertTime,
		&post.LastUpdateTime,
	)
	if err != nil {
		return post, translateError(err)
	}
	post.ID = model.MakePostID(dbID)
	post.Collection = model.MakeCollectionID(catID)
	return post, nil
}

// InsertPost inserts a PostBody into the database and returns the inserted Post
func (p PostgresPersister) InsertPost(ctx context.Context, in model.PostIn) (model.Post, error) {
	var dbID int32
	var post model.Post
	var catID int32
	n, err := fmt.Sscanf(in.Collection+"\n", p.pathPrefix+model.CollectionIDFormat+"\n", &catID)
	if err != nil || n != 1 {
		// Bad ID
		return post, model.NewError(model.ErrBadReference, in.Collection, "collection")
	}
	err = p.db.QueryRowContext(ctx,
		`INSERT INTO post (collection_id, body) 
		 VALUES ($1, $2) 
		 RETURNING id, collection_id, body, insert_time, last_update_time`,
		catID, in.PostBody).
		Scan(
			&dbID,
			&catID,
			&post.PostBody,
			&post.InsertTime,
			&post.LastUpdateTime,
		)
	post.ID = model.MakePostID(dbID)
	post.Collection = model.MakeCollectionID(catID)
	return post, translateError(err)
}

// UpdatePost updates a Post in the database and returns the updated Post
func (p PostgresPersister) UpdatePost(ctx context.Context, id string, in model.Post) (model.Post, error) {
	var post model.Post
	var dbID int32
	n, err := fmt.Sscanf(id+"\n", p.pathPrefix+model.PostIDFormat+"\n", &dbID)
	if err != nil || n != 1 {
		// Bad ID
		return post, model.NewError(model.ErrNotFound, id)
	}
	var catID int32
	n, err = fmt.Sscanf(in.Collection+"\n", p.pathPrefix+model.CollectionIDFormat+"\n", &catID)
	if err != nil || n != 1 {
		// Bad ID
		return post, model.NewError(model.ErrBadReference, in.Collection, "collection")
	}
	err = p.db.QueryRowContext(ctx,
		`UPDATE post SET body = $1, collection_id = $2, last_update_time = CURRENT_TIMESTAMP 
		 WHERE id = $3 AND last_update_time = $4
		 RETURNING id, collection_id, body, insert_time, last_update_time`,
		in.PostBody, catID, dbID, in.LastUpdateTime).
		Scan(
			&dbID,
			&catID,
			&post.PostBody,
			&post.InsertTime,
			&post.LastUpdateTime,
		)
	if err != nil && err == sql.ErrNoRows {
		// Either non-existent or last_update_time didn't match
		c, _ := p.SelectOnePost(ctx, id)
		if c.ID == id {
			// Row exists, so it must be a non-matching update time
			return post, model.NewError(model.ErrConcurrentUpdate, c.LastUpdateTime.String(), in.LastUpdateTime.String())
		}
		return post, model.NewError(model.ErrNotFound, id)
	}
	post.ID = model.MakePostID(dbID)
	post.Collection = model.MakeCollectionID(catID)
	return post, translateError(err)
}

// DeletePost deletes a Post
func (p PostgresPersister) DeletePost(ctx context.Context, id string) error {
	var dbID int32
	n, err := fmt.Sscanf(id+"\n", p.pathPrefix+model.PostIDFormat+"\n", &dbID)
	if err != nil || n != 1 {
		// Bad ID
		return model.NewError(model.ErrNotFound, id)
	}
	_, err = p.db.ExecContext(ctx, "DELETE FROM post WHERE id = $1", dbID)
	return translateError(err)
}
