package main

import (
	"context"

	"github.com/ourrootsorg/cms-server/api"
	"github.com/ourrootsorg/cms-server/model"
)

type apiMock struct {
	// mock.Mock
	result interface{}
	errors *model.Errors
}

func (a *apiMock) GetCategories(ctx context.Context) (*api.CategoryResult, *model.Errors) {
	// a.Called(ctx)
	return a.result.(*api.CategoryResult), a.errors
}
func (a *apiMock) GetCategory(ctx context.Context, id string) (*model.Category, *model.Errors) {
	return a.result.(*model.Category), a.errors
}
func (a *apiMock) AddCategory(ctx context.Context, in model.CategoryIn) (*model.Category, *model.Errors) {
	return a.result.(*model.Category), a.errors
}
func (a *apiMock) UpdateCategory(ctx context.Context, id string, in model.Category) (*model.Category, *model.Errors) {
	return a.result.(*model.Category), a.errors
}
func (a *apiMock) DeleteCategory(ctx context.Context, id string) *model.Errors {
	return a.errors
}
func (a *apiMock) GetCollections(ctx context.Context /* filter/search criteria */) (*api.CollectionResult, *model.Errors) {
	return a.result.(*api.CollectionResult), a.errors
}
func (a *apiMock) GetCollection(ctx context.Context, id string) (*model.Collection, *model.Errors) {
	return a.result.(*model.Collection), a.errors
}
func (a *apiMock) AddCollection(ctx context.Context, in model.CollectionIn) (*model.Collection, *model.Errors) {
	return a.result.(*model.Collection), a.errors
}
func (a *apiMock) UpdateCollection(ctx context.Context, id string, in model.Collection) (*model.Collection, *model.Errors) {
	return a.result.(*model.Collection), a.errors
}
func (a *apiMock) DeleteCollection(ctx context.Context, id string) *model.Errors {
	return a.errors
}
