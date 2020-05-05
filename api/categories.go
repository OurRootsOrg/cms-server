package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist"
)

// CategoryResult is a paged Category result
type CategoryResult struct {
	Categories []model.Category `json:"collections"`
	NextPage   string           `json:"next_page"`
}

// GetCategories holds the business logic around getting many Categories
func (api API) GetCategories( /* filter/search criteria */ ) (*CategoryResult, *Errors) {
	// TODO: handle search criteria and paged results
	cols, err := api.categoryPersister.SelectCategories()
	if err != nil {
		return nil, NewErrors(http.StatusInternalServerError, err)
	}
	return &CategoryResult{Categories: cols}, nil
}

// GetCategory holds the business logic around getting a Category
func (api API) GetCategory(id string) (*model.Category, *Errors) {
	collection, err := api.categoryPersister.SelectOneCategory(id)
	if err == persist.ErrNoRows {
		msg := fmt.Sprintf("Not Found: %v", err)
		log.Print("[ERROR] " + msg)
		return nil, NewErrors(http.StatusNotFound, err)
	} else if err != nil {
		return nil, NewErrors(http.StatusInternalServerError, err)
	}
	return &collection, nil
}

// AddCategory holds the business logic around adding a Category
func (api API) AddCategory(in model.CategoryIn) (*model.Category, *Errors) {
	err := api.validate.Struct(in)
	if err != nil {
		return nil, NewErrors(http.StatusBadRequest, err)
	}
	collection, err := api.categoryPersister.InsertCategory(in)
	if err != nil {
		return nil, NewErrors(http.StatusInternalServerError, err)
	}
	return &collection, nil
}

// UpdateCategory holds the business logic around updating a Category
func (api API) UpdateCategory(id string, in model.CategoryIn) (*model.Category, *Errors) {
	err := api.validate.Struct(in)
	if err != nil {
		return nil, NewErrors(http.StatusBadRequest, err)
	}
	collection, err := api.categoryPersister.UpdateCategory(id, in)
	if err == persist.ErrNoRows {
		// Not allowed to add a Category with PUT
		return nil, NewErrors(http.StatusNotFound, NewError(ErrNotFound, "collection"))
	} else if err != nil {
		return nil, NewErrors(http.StatusInternalServerError, err)
	}
	return &collection, nil
}

// DeleteCategory holds the business logic around deleting a Category
func (api API) DeleteCategory(id string) *Errors {
	err := api.categoryPersister.DeleteCategory(id)
	if err != nil {
		return NewErrors(http.StatusInternalServerError, err)
	}
	return nil
}
