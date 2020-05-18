package persist

import (
	"context"
	"math/rand"
	"time"

	"github.com/ourrootsorg/cms-server/model"
)

// MemoryPersister "persists" the model objects to an in-memory map.
// It's mostly useful for testing
type MemoryPersister struct {
	pathPrefix  string
	categories  map[string]model.Category
	collections map[string]model.Collection
	posts       map[string]model.Post
}

// NewMemoryPersister constructs a MemoryPersister
func NewMemoryPersister(pathPrefix string) MemoryPersister {
	return MemoryPersister{
		pathPrefix:  pathPrefix,
		categories:  make(map[string]model.Category, 0),
		collections: make(map[string]model.Collection, 0),
	}
}

// Category persistence methods

// SelectCategories loads all the categories from the database
func (p MemoryPersister) SelectCategories(ctx context.Context) ([]model.Category, error) {
	cats := make([]model.Category, 0, len(p.categories))

	for _, value := range p.categories {
		cats = append(cats, value)
	}
	return cats, nil
}

// SelectOneCategory loads a single category from the database
func (p MemoryPersister) SelectOneCategory(ctx context.Context, id string) (model.Category, error) {
	cat, found := p.categories[id]
	if !found {
		return cat, ErrNoRows
	}
	return cat, nil
}

// InsertCategory inserts a CategoryBody into the database and returns the inserted Category
func (p MemoryPersister) InsertCategory(ctx context.Context, in model.CategoryIn) (model.Category, error) {
	cat := model.NewCategory(int32(rand.Int31()), in)
	now := time.Now()
	cat.InsertTime = now
	cat.LastUpdateTime = now
	// Add to "database"
	p.categories[string(cat.ID)] = cat
	return cat, nil
}

// UpdateCategory updates a Category in the database and returns the updated Category
func (p MemoryPersister) UpdateCategory(ctx context.Context, id string, in model.Category) (model.Category, error) {
	_, found := p.categories[id]
	if !found {
		return model.Category{}, ErrNoRows
	}
	in.LastUpdateTime = time.Now()
	p.categories[string(in.ID)] = in
	return in, nil
}

// DeleteCategory deletes a Category
func (p MemoryPersister) DeleteCategory(ctx context.Context, id string) error {
	delete(p.categories, id)
	return nil
}

// Collection persistence methods

// SelectCollections selects all collections
func (p MemoryPersister) SelectCollections(ctx context.Context) ([]model.Collection, error) {
	cols := make([]model.Collection, 0, len(p.collections))

	for _, value := range p.collections {
		cols = append(cols, value)
	}
	return cols, nil
}

// SelectOneCollection selects a single collection
func (p MemoryPersister) SelectOneCollection(ctx context.Context, id string) (model.Collection, error) {
	col, found := p.collections[id]
	if !found {
		return col, ErrNoRows
	}
	return col, nil
}

// InsertCollection inserts a new collection
func (p MemoryPersister) InsertCollection(ctx context.Context, in model.CollectionIn) (model.Collection, error) {
	col := model.NewCollection(int32(rand.Int31()), in)
	now := time.Now()
	col.InsertTime = now
	col.LastUpdateTime = now
	// Add to "database"
	p.collections[col.ID] = col
	return col, nil
}

// UpdateCollection updates a collection
func (p MemoryPersister) UpdateCollection(ctx context.Context, id string, in model.Collection) (model.Collection, error) {
	col := model.Collection{}
	_, found := p.collections[id]
	if !found {
		return col, ErrNoRows
	}
	col.ID = id
	col.CollectionBody = in.CollectionBody
	col.Category = in.Category
	now := time.Now()
	col.LastUpdateTime = now
	p.collections[col.ID] = col
	return col, nil
}

// DeleteCollection deletes a collection
func (p MemoryPersister) DeleteCollection(ctx context.Context, id string) error {
	delete(p.collections, id)
	return nil
}

// Post persistence methods

// SelectPosts selects all posts
func (p MemoryPersister) SelectPosts(ctx context.Context) ([]model.Post, error) {
	cols := make([]model.Post, 0, len(p.posts))

	for _, value := range p.posts {
		cols = append(cols, value)
	}
	return cols, nil
}

// SelectOnePost selects a single post
func (p MemoryPersister) SelectOnePost(ctx context.Context, id string) (model.Post, error) {
	col, found := p.posts[id]
	if !found {
		return col, ErrNoRows
	}
	return col, nil
}

// InsertPost inserts a new post
func (p MemoryPersister) InsertPost(ctx context.Context, in model.PostIn) (model.Post, error) {
	col := model.NewPost(int32(rand.Int31()), in)
	now := time.Now()
	col.InsertTime = now
	col.LastUpdateTime = now
	// Add to "database"
	p.posts[col.ID] = col
	return col, nil
}

// UpdatePost updates a post
func (p MemoryPersister) UpdatePost(ctx context.Context, id string, in model.Post) (model.Post, error) {
	col := model.Post{}
	_, found := p.posts[id]
	if !found {
		return col, ErrNoRows
	}
	col.ID = id
	col.PostBody = in.PostBody
	col.Collection = in.Collection
	now := time.Now()
	col.LastUpdateTime = now
	p.posts[col.ID] = col
	return col, nil
}

// DeletePost deletes a post
func (p MemoryPersister) DeletePost(ctx context.Context, id string) error {
	delete(p.posts, id)
	return nil
}
