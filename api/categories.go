package api

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist"
)

// CategoryResult is a paged Category result
type CategoryResult struct {
	Categories []model.Category `json:"categories"`
	NextPage   string           `json:"next_page"`
}

// GetCategories holds the business logic around getting many Categories
func (api API) GetCategories(ctx context.Context /* filter/search criteria */) (*CategoryResult, *model.Errors) {
	// TODO: handle search criteria and paged results
	cols, err := api.categoryPersister.SelectCategories(ctx)
	if err != nil {
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	return &CategoryResult{Categories: cols}, nil
}

// GetCategory holds the business logic around getting a Category
func (api API) GetCategory(ctx context.Context, id string) (*model.Category, *model.Errors) {
	category, err := api.categoryPersister.SelectOneCategory(ctx, id)
	if err == persist.ErrNoRows {
		msg := fmt.Sprintf("Not Found: %v", err)
		log.Print("[ERROR] " + msg)
		return nil, model.NewErrors(http.StatusNotFound, model.NewError(model.ErrNotFound, id))
	} else if err != nil {
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	return &category, nil
}

// AddCategory holds the business logic around adding a Category
func (api API) AddCategory(ctx context.Context, in model.CategoryIn) (*model.Category, *model.Errors) {
	err := api.validate.Struct(in)
	if err != nil {
		return nil, model.NewErrors(http.StatusBadRequest, err)
	}
	category, err := api.categoryPersister.InsertCategory(ctx, in)
	if err != nil {
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	return &category, nil
}

// UpdateCategory holds the business logic around updating a Category
func (api API) UpdateCategory(ctx context.Context, id string, in model.Category) (*model.Category, *model.Errors) {
	err := api.validate.Struct(in)
	if err != nil {
		return nil, model.NewErrors(http.StatusBadRequest, err)
	}
	category, err := api.categoryPersister.UpdateCategory(ctx, id, in)
	if er, ok := err.(model.Error); ok {
		if er.Code == model.ErrConcurrentUpdate {
			return nil, model.NewErrors(http.StatusConflict, er)
		} else if er.Code == model.ErrNotFound {
			// Not allowed to add a Category with PUT
			return nil, model.NewErrors(http.StatusNotFound, er)
		}
	}
	if err != nil {
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	return &category, nil
}

// DeleteCategory holds the business logic around deleting a Category
func (api API) DeleteCategory(ctx context.Context, id string) *model.Errors {
	err := api.categoryPersister.DeleteCategory(ctx, id)
	if err != nil {
		return model.NewErrors(http.StatusInternalServerError, err)
	}
	return nil
}
