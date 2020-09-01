package api

import (
	"context"

	"github.com/ourrootsorg/cms-server/model"
)

func (api *API) GetNameVariants(ctx context.Context, nameType model.NameType, name string) (*model.NameVariants, error) {
	return api.namePersister.SelectNameVariants(ctx, nameType, name)
}
