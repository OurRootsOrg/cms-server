package api

import (
	"context"
	"net/http"

	"github.com/ourrootsorg/cms-server/model"
)

// GetSettings holds the business logic around getting a Settings object
func (api API) GetSettings(ctx context.Context) (*model.Settings, *model.Errors) {
	settings, err := api.settingsPersister.SelectSettings(ctx)
	if err != nil {
		// if no settings, return a default settings object
		if err.Code == model.ErrNotFound {
			settings = &model.Settings{
				SettingsIn: model.NewSettingsIn([]model.SettingsPostMetadata{}),
			}
			err = nil
		} else {
			return nil, model.NewErrorsFromError(err)
		}
	}
	return settings, nil
}

// UpdateSettings holds the business logic around updating a Settings object
func (api API) UpdateSettings(ctx context.Context, in model.Settings) (*model.Settings, *model.Errors) {
	err := api.validate.Struct(in)
	if err != nil {
		return nil, model.NewErrors(http.StatusBadRequest, err)
	}

	settings, e := api.settingsPersister.UpsertSettings(ctx, in)
	if e != nil {
		return nil, model.NewErrorsFromError(e)
	}
	return settings, nil
}
