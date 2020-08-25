package api

import (
	"context"

	"github.com/ourrootsorg/cms-server/model"
)

func (api *API) StandardizePlace(ctx context.Context, text, defaultContainingPlace string) (*model.Place, error) {
	return api.placeStandardizer.Standardize(ctx, text, defaultContainingPlace)
}

func (api *API) GetPlacesByPrefix(ctx context.Context, prefix string, count int) ([]model.Place, error) {
	return api.placePersister.SelectPlacesByFullNamePrefix(ctx, prefix, count)
}
