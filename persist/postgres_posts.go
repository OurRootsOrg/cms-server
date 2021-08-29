package persist

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/ourrootsorg/cms-server/utils"

	"github.com/ourrootsorg/cms-server/model"
)

// SelectPosts selects all posts
func (p PostgresPersister) SelectPosts(ctx context.Context) ([]model.Post, error) {
	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := p.db.QueryContext(ctx, "SELECT id, collection_id, body, insert_time, last_update_time FROM post "+
		"WHERE society_id=$1", societyID)
	if err != nil {
		return nil, translateError(err, nil, nil, "")
	}
	defer rows.Close()
	posts := make([]model.Post, 0)
	for rows.Next() {
		var post model.Post
		err := rows.Scan(&post.ID, &post.Collection, &post.PostBody, &post.InsertTime, &post.LastUpdateTime)
		if err != nil {
			return nil, translateError(err, nil, nil, "")
		}
		posts = append(posts, post)
	}
	return posts, nil
}

// SelectOnePost selects a single post
func (p PostgresPersister) SelectOnePost(ctx context.Context, id uint32) (*model.Post, error) {
	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var post model.Post
	err = p.db.QueryRowContext(ctx, "SELECT id, collection_id, body, insert_time, last_update_time FROM post "+
		"WHERE society_id=$1 AND id=$2", societyID, id).Scan(
		&post.ID,
		&post.Collection,
		&post.PostBody,
		&post.InsertTime,
		&post.LastUpdateTime,
	)
	if err != nil {
		return nil, translateError(err, &id, nil, "")
	}
	return &post, nil
}

// InsertPost inserts a PostBody into the database and returns the inserted Post
func (p PostgresPersister) InsertPost(ctx context.Context, in model.PostIn) (*model.Post, error) {
	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var post model.Post
	err = p.db.QueryRowContext(ctx,
		`INSERT INTO post (society_id, collection_id, body)
		 VALUES ($1, $2, $3)
		 RETURNING id, collection_id, body, insert_time, last_update_time`,
		societyID, in.Collection, in.PostBody).
		Scan(
			&post.ID,
			&post.Collection,
			&post.PostBody,
			&post.InsertTime,
			&post.LastUpdateTime,
		)
	return &post, translateError(err, nil, &in.Collection, "collection")
}

// UpdatePost updates a Post in the database and returns the updated Post
func (p PostgresPersister) UpdatePost(ctx context.Context, id uint32, in model.Post) (*model.Post, error) {
	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var post model.Post
	err = p.db.QueryRowContext(ctx,
		`UPDATE post SET body = $1, collection_id = $2, last_update_time = CURRENT_TIMESTAMP
		 WHERE society_id=$3 AND id = $4 AND last_update_time = $5
		 RETURNING id, collection_id, body, insert_time, last_update_time`,
		in.PostBody, in.Collection, societyID, id, in.LastUpdateTime).
		Scan(
			&post.ID,
			&post.Collection,
			&post.PostBody,
			&post.InsertTime,
			&post.LastUpdateTime,
		)
	if err != nil && err == sql.ErrNoRows {
		// Either non-existent or last_update_time didn't match
		c, _ := p.SelectOnePost(ctx, id)
		if c.ID == id {
			// Row exists, so it must be a non-matching update time
			return nil, model.NewError(model.ErrConcurrentUpdate, c.LastUpdateTime.String(), in.LastUpdateTime.String())
		}
		return nil, model.NewError(model.ErrNotFound, strconv.Itoa(int(id)))
	}
	return &post, translateError(err, &id, &in.Collection, "collection")
}

// DeletePost deletes a Post
func (p PostgresPersister) DeletePost(ctx context.Context, id uint32) error {
	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		return err
	}
	_, err = p.db.ExecContext(ctx, "DELETE FROM post WHERE society_id=$1 AND id = $2", societyID, id)
	return translateError(err, nil, nil, "")
}
