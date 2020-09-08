package api

import (
	"context"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/go-oidc"
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
func (a *ApiMock) GetPostImage(ctx context.Context, id uint32, filePath string, thumbnail bool, expireSeconds int) (*ImageMetadata, error) {
	return a.Result.(*ImageMetadata), a.Errors
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

func (a *ApiMock) GetRecordsForPost(ctx context.Context, postid uint32) (*RecordsResult, error) {
	return a.Result.(*RecordsResult), a.Errors
}
func (a *ApiMock) GetRecordsByID(ctx context.Context, ids []uint32) ([]model.Record, error) {
	return a.Result.([]model.Record), a.Errors
}
func (a *ApiMock) GetRecord(ctx context.Context, includeDetails bool, id uint32) (*RecordDetail, error) {
	return a.Result.(*RecordDetail), a.Errors
}
func (a *ApiMock) AddRecord(ctx context.Context, in model.RecordIn) (*model.Record, error) {
	return a.Result.(*model.Record), a.Errors
}
func (a *ApiMock) UpdateRecord(ctx context.Context, id uint32, in model.Record) (*model.Record, error) {
	return a.Result.(*model.Record), a.Errors
}
func (a *ApiMock) DeleteRecord(ctx context.Context, id uint32) error {
	return a.Errors
}
func (a *ApiMock) DeleteRecordsForPost(ctx context.Context, postID uint32) error {
	return a.Errors
}
func (a *ApiMock) GetRecordHouseholdsForPost(ctx context.Context, postid uint32) ([]model.RecordHousehold, error) {
	return a.Result.([]model.RecordHousehold), a.Errors
}
func (a *ApiMock) GetRecordHousehold(ctx context.Context, postID uint32, householdID string) (*model.RecordHousehold, error) {
	return a.Result.(*model.RecordHousehold), a.Errors
}
func (a *ApiMock) AddRecordHousehold(ctx context.Context, in model.RecordHouseholdIn) (*model.RecordHousehold, error) {
	return a.Result.(*model.RecordHousehold), a.Errors
}
func (a *ApiMock) DeleteRecordHouseholdsForPost(ctx context.Context, postID uint32) error {
	return a.Errors
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

func (a *ApiMock) StandardizePlace(ctx context.Context, text, defaultContainingPlace string) (*model.Place, error) {
	return a.Result.(*model.Place), a.Errors
}
func (a *ApiMock) GetPlacesByPrefix(ctx context.Context, prefix string, count int) ([]model.Place, error) {
	a.Request = prefix
	return a.Result.([]model.Place), a.Errors
}

func (a *ApiMock) GetNameVariants(ctx context.Context, nameType model.NameType, name string) (*model.NameVariants, error) {
	return a.Result.(*model.NameVariants), a.Errors
}
