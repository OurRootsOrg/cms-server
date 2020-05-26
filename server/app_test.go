package main

import (
	"context"

	"github.com/coreos/go-oidc"
	"github.com/ourrootsorg/cms-server/api"
	"github.com/ourrootsorg/cms-server/model"
)

type apiMock struct {
	// mock.Mock
	result interface{}
	errors *model.Errors
}

func (a *apiMock) GetCategories(ctx context.Context) (*api.CategoryResult, *model.Errors) {
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

func (a *apiMock) GetPosts(ctx context.Context /* filter/search criteria */) (*api.PostResult, *model.Errors) {
	return a.result.(*api.PostResult), a.errors
}
func (a *apiMock) GetPost(ctx context.Context, id string) (*model.Post, *model.Errors) {
	return a.result.(*model.Post), a.errors
}
func (a *apiMock) AddPost(ctx context.Context, in model.PostIn) (*model.Post, *model.Errors) {
	return a.result.(*model.Post), a.errors
}
func (a *apiMock) UpdatePost(ctx context.Context, id string, in model.Post) (*model.Post, *model.Errors) {
	return a.result.(*model.Post), a.errors
}
func (a *apiMock) DeletePost(ctx context.Context, id string) *model.Errors {
	return a.errors
}
func (a *apiMock) PostContentRequest(ctx context.Context, contentRequest api.ContentRequest) (*api.ContentResult, *model.Errors) {
	return a.result.(*api.ContentResult), a.errors
}
func (a *apiMock) GetContent(ctx context.Context, key string) ([]byte, *model.Errors) {
	return a.result.([]byte), a.errors
}
func (a *apiMock) RetrieveUser(ctx context.Context, provider api.OIDCProvider, token *oidc.IDToken, rawToken string) (*model.User, *model.Errors) {
	return a.result.(*model.User), a.errors
}
func (a *apiMock) Search(ctx context.Context, searchRequest api.SearchRequest) (*api.SearchResult, *model.Errors) {
	return a.result.(*api.SearchResult), a.errors
}
