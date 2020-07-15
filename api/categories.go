package api

import (
	"context"

	"github.com/ourrootsorg/cms-server/model"
)

// CategoryResult is a paged Category result
type CategoryResult struct {
	Categories []model.Category `json:"categories"`
	NextPage   string           `json:"next_page"`
}

// GetCategories holds the business logic around getting many Categories
func (api API) GetCategories(ctx context.Context /* filter/search criteria */) (*CategoryResult, error) {
	// TODO: handle search criteria and paged results
	cols, err := api.categoryPersister.SelectCategories(ctx)
	if err != nil {
		return nil, NewError(err)
	}
	return &CategoryResult{Categories: cols}, nil
}

// GetCategoriesByID holds the business logic around getting many Categories
func (api API) GetCategoriesByID(ctx context.Context, ids []uint32) ([]model.Category, error) {
	cats, err := api.categoryPersister.SelectCategoriesByID(ctx, ids)
	if err != nil {
		return nil, NewError(err)
	}
	return cats, nil
}

// GetCategory holds the business logic around getting a Category
func (api API) GetCategory(ctx context.Context, id uint32) (*model.Category, error) {
	category, err := api.categoryPersister.SelectOneCategory(ctx, id)
	if err != nil {
		return nil, NewError(err)
	}
	return category, nil
}

// AddCategory holds the business logic around adding a Category
func (api API) AddCategory(ctx context.Context, in model.CategoryIn) (*model.Category, error) {
	err := api.validate.Struct(in)
	if err != nil {
		return nil, NewError(err)
	}
	category, e := api.categoryPersister.InsertCategory(ctx, in)
	if err != nil {
		return nil, NewError(e)
	}
	return category, nil
}

// UpdateCategory holds the business logic around updating a Category
func (api API) UpdateCategory(ctx context.Context, id uint32, in model.Category) (*model.Category, error) {
	err := api.validate.Struct(in)
	if err != nil {
		return nil, NewError(err)
	}
	category, err := api.categoryPersister.UpdateCategory(ctx, id, in)
	if err != nil {
		return nil, NewError(err)
	}
	return category, nil
}

// DeleteCategory holds the business logic around deleting a Category
func (api API) DeleteCategory(ctx context.Context, id uint32) error {
	err := api.categoryPersister.DeleteCategory(ctx, id)
	if err != nil {
		return NewError(err)
	}
	return nil
}
