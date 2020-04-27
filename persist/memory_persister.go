package persist

import (
	"math/rand"
	"time"

	"github.com/jancona/ourroots/model"
)

// MemoryPersister "persists" the model objects to an in-memory map.
// It's mostly useful for testing
type MemoryPersister struct {
	pathPrefix  string
	categories  map[string]model.Category
	collections map[string]model.Collection
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
func (p MemoryPersister) SelectCategories() ([]model.Category, error) {
	cats := make([]model.Category, 0, len(p.categories))

	for _, value := range p.categories {
		cats = append(cats, value)
	}
	return cats, nil
}

// SelectOneCategory loads a single category from the database
func (p MemoryPersister) SelectOneCategory(id string) (model.Category, error) {
	cat, found := p.categories[id]
	if !found {
		return cat, ErrNoRows
	}
	return cat, nil
}

// InsertCategory inserts a CategoryBody into the database and returns the inserted Category
func (p MemoryPersister) InsertCategory(in model.CategoryIn) (model.Category, error) {
	cat := model.NewCategory(int32(rand.Int31()), in)
	cat.Type = "category"
	now := time.Now()
	cat.InsertTime = now
	cat.LastUpdateTime = now
	// Add to "database"
	p.categories[cat.ID] = cat
	return cat, nil
}

// UpdateCategory updates a Category in the database and returns the updated Category
func (p MemoryPersister) UpdateCategory(id string, in model.CategoryIn) (model.Category, error) {
	_, found := p.categories[id]
	if !found {
		return model.Category{}, ErrNoRows
	}
	cat := model.NewCategory(0, in)
	cat.ID = id
	cat.LastUpdateTime = time.Now()
	p.categories[cat.ID] = cat
	return cat, nil
}

// DeleteCategory deletes a Category
func (p MemoryPersister) DeleteCategory(id string) error {
	delete(p.categories, id)
	return nil
}

// Collection persistence methods

// SelectCollections selects all collections
func (p MemoryPersister) SelectCollections() ([]model.Collection, error) {
	cols := make([]model.Collection, 0, len(p.collections))

	for _, value := range p.collections {
		cols = append(cols, value)
	}
	return cols, nil
}

// SelectOneCollection selects a single collection
func (p MemoryPersister) SelectOneCollection(id string) (model.Collection, error) {
	col, found := p.collections[id]
	if !found {
		return col, ErrNoRows
	}
	return col, nil
}

// InsertCollection inserts a new collection
func (p MemoryPersister) InsertCollection(in model.CollectionIn) (model.Collection, error) {
	col := model.NewCollection(int32(rand.Int31()), in)
	col.Type = "collection"
	now := time.Now()
	col.InsertTime = now
	col.LastUpdateTime = now
	// Add to "database"
	p.collections[col.ID] = col
	return col, nil
}

// UpdateCollection updates a collection
func (p MemoryPersister) UpdateCollection(id string, in model.CollectionIn) (model.Collection, error) {
	col := model.Collection{}
	_, found := p.collections[id]
	if !found {
		return col, ErrNoRows
	}
	col.ID = id
	col.Type = "collection"
	col.CollectionBody = in.CollectionBody
	col.Category = in.Category
	col.LastUpdateTime = time.Now()
	p.collections[col.ID] = col
	return col, nil
}

// DeleteCollection deletes a collection
func (p MemoryPersister) DeleteCollection(id string) error {
	delete(p.collections, id)
	return nil
}
