package api

import (
	"context"

	"github.com/coreos/go-oidc"
	"github.com/ourrootsorg/cms-server/model"
)

type ApiMock struct {
	// mock.Mock
	Request interface{}
	Result  interface{}
	Errors  error
}

func (a *ApiMock) GetCategories(ctx context.Context) (*CategoryResult, error) {
	return a.Result.(*CategoryResult), a.Errors
}
func (a *ApiMock) GetCategoriesByID(ctx context.Context, ids []uint32) ([]model.Category, error) {
	return a.Result.([]model.Category), a.Errors
}
func (a *ApiMock) GetCategory(ctx context.Context, id uint32) (*model.Category, error) {
	return a.Result.(*model.Category), a.Errors
}
func (a *ApiMock) AddCategory(ctx context.Context, in model.CategoryIn) (*model.Category, error) {
	return a.Result.(*model.Category), a.Errors
}
func (a *ApiMock) UpdateCategory(ctx context.Context, id uint32, in model.Category) (*model.Category, error) {
	return a.Result.(*model.Category), a.Errors
}
func (a *ApiMock) DeleteCategory(ctx context.Context, id uint32) error {
	return a.Errors
}
func (a *ApiMock) GetCollections(ctx context.Context /* filter/search criteria */) (*CollectionResult, error) {
	return a.Result.(*CollectionResult), a.Errors
}
func (a *ApiMock) GetCollectionsByID(ctx context.Context, ids []uint32) ([]model.Collection, error) {
	return a.Result.([]model.Collection), a.Errors
}
func (a *ApiMock) GetCollection(ctx context.Context, id uint32) (*model.Collection, error) {
	return a.Result.(*model.Collection), a.Errors
}
func (a *ApiMock) AddCollection(ctx context.Context, in model.CollectionIn) (*model.Collection, error) {
	return a.Result.(*model.Collection), a.Errors
}
func (a *ApiMock) UpdateCollection(ctx context.Context, id uint32, in model.Collection) (*model.Collection, error) {
	return a.Result.(*model.Collection), a.Errors
}
func (a *ApiMock) DeleteCollection(ctx context.Context, id uint32) error {
	return a.Errors
}

func (a *ApiMock) GetPosts(ctx context.Context /* filter/search criteria */) (*PostResult, error) {
	return a.Result.(*PostResult), a.Errors
}
func (a *ApiMock) GetPost(ctx context.Context, id uint32) (*model.Post, error) {
	return a.Result.(*model.Post), a.Errors
}
func (a *ApiMock) AddPost(ctx context.Context, in model.PostIn) (*model.Post, error) {
	return a.Result.(*model.Post), a.Errors
}
func (a *ApiMock) UpdatePost(ctx context.Context, id uint32, in model.Post) (*model.Post, error) {
	return a.Result.(*model.Post), a.Errors
}
func (a *ApiMock) DeletePost(ctx context.Context, id uint32) error {
	return a.Errors
}

func (a *ApiMock) PostContentRequest(ctx context.Context, contentRequest ContentRequest) (*ContentResult, error) {
	return a.Result.(*ContentResult), a.Errors
}
func (a *ApiMock) GetContent(ctx context.Context, key string) ([]byte, error) {
	return a.Result.([]byte), a.Errors
}

func (a *ApiMock) RetrieveUser(ctx context.Context, provider OIDCProvider, token *oidc.IDToken, rawToken string) (*model.User, error) {
	return a.Result.(*model.User), a.Errors
}

func (a *ApiMock) GetRecordsForPost(ctx context.Context, postid uint32) (*RecordResult, error) {
	return a.Result.(*RecordResult), a.Errors
}

func (a *ApiMock) Search(ctx context.Context, searchRequest *SearchRequest) (*model.SearchResult, error) {
	a.Request = searchRequest
	return a.Result.(*model.SearchResult), a.Errors
}
func (a *ApiMock) SearchByID(ctx context.Context, id string) (*model.SearchHit, error) {
	return a.Result.(*model.SearchHit), a.Errors
}
func (a *ApiMock) SearchDeleteByID(ctx context.Context, id string) error {
	return a.Errors
}

func (a *ApiMock) GetSettings(ctx context.Context) (*model.Settings, error) {
	return a.Result.(*model.Settings), a.Errors
}
func (a *ApiMock) UpdateSettings(ctx context.Context, in model.Settings) (*model.Settings, error) {
	return a.Result.(*model.Settings), a.Errors
}
