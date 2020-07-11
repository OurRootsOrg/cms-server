package api

import (
	"context"
	"net/http"

	"github.com/ourrootsorg/cms-server/model"
)

// CategoryResult is a paged Category result
type CategoryResult struct {
	Categories []model.Category `json:"categories"`
	NextPage   string           `json:"next_page"`
}

// GetCategories holds the business logic around getting many Categories
func (api API) GetCategories(ctx context.Context /* filter/search criteria */) (*CategoryResult, *model.Errors) {
	// token, ok := ctx.Value(TokenProperty).(*oidc.IDToken)
	// if token == nil || !ok {
	// 	msg := "No authentication token"
	// 	log.Print("[DEBUG] " + msg)
	// 	return nil, model.NewErrors(http.StatusUnauthorized, errors.New(msg))
	// }
	// log.Printf("[DEBUG] This is an authenticated request")
	// log.Printf("[DEBUG] Audience: %s", token.Audience)
	// log.Printf("[DEBUG] Subject: %s", token.Subject)
	// log.Printf("[DEBUG] Issuer: %s", token.Issuer)
	// log.Printf("[DEBUG] IssuedAt: %s", token.IssuedAt.Format(time.RFC3339))
	// log.Printf("[DEBUG] Expiry: %s", token.Expiry.Format(time.RFC3339))
	// claims := make(map[string]interface{})
	// err := token.Claims(&claims)
	// if err != nil {
	// 	log.Printf("[ERROR] Error getting claims: %v", err)
	// }
	// log.Printf("[DEBUG] Claims: %#v", claims)
	// TODO: handle search criteria and paged results
	cols, err := api.categoryPersister.SelectCategories(ctx)
	if err != nil {
		return nil, model.NewErrorsFromError(err)
	}
	return &CategoryResult{Categories: cols}, nil
}

// GetCategoriesByID holds the business logic around getting many Categories
func (api API) GetCategoriesByID(ctx context.Context, ids []uint32) ([]model.Category, *model.Errors) {
	cats, err := api.categoryPersister.SelectCategoriesByID(ctx, ids)
	if err != nil {
		return nil, model.NewErrorsFromError(err)
	}
	return cats, nil
}

// GetCategory holds the business logic around getting a Category
func (api API) GetCategory(ctx context.Context, id uint32) (*model.Category, *model.Errors) {
	category, err := api.categoryPersister.SelectOneCategory(ctx, id)
	if err != nil {
		return nil, model.NewErrorsFromError(err)
	}
	return category, nil
}

// AddCategory holds the business logic around adding a Category
func (api API) AddCategory(ctx context.Context, in model.CategoryIn) (*model.Category, *model.Errors) {
	err := api.validate.Struct(in)
	if err != nil {
		return nil, model.NewErrors(http.StatusBadRequest, err)
	}
	category, e := api.categoryPersister.InsertCategory(ctx, in)
	if err != nil {
		return nil, model.NewErrorsFromError(e)
	}
	return category, nil
}

// UpdateCategory holds the business logic around updating a Category
func (api API) UpdateCategory(ctx context.Context, id uint32, in model.Category) (*model.Category, *model.Errors) {
	err := api.validate.Struct(in)
	if err != nil {
		return nil, model.NewErrors(http.StatusBadRequest, err)
	}
	category, e := api.categoryPersister.UpdateCategory(ctx, id, in)
	if e != nil {
		return nil, model.NewErrorsFromError(e)
	}
	return category, nil
}

// DeleteCategory holds the business logic around deleting a Category
func (api API) DeleteCategory(ctx context.Context, id uint32) *model.Errors {
	err := api.categoryPersister.DeleteCategory(ctx, id)
	if err != nil {
		return model.NewErrorsFromError(err)
	}
	return nil
}
