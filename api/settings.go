package api

import (
	"context"
	"net/http"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist"
)

// GetSettings holds the business logic around getting a Settings object
func (api API) GetSettings(ctx context.Context) (*model.Settings, *model.Errors) {
	settings, err := api.settingsPersister.SelectSettings(ctx)
	// if no settings, return a default settings object
	if err == persist.ErrNoRows {
		settings = model.Settings{
			SettingsIn: model.NewSettingsIn([]model.SettingsPostField{}),
		}
		err = nil
	} else if err != nil {
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	return &settings, nil
}

// UpsertSettings holds the business logic around updating a Settings object
func (api API) UpdateSettings(ctx context.Context, in model.Settings) (*model.Settings, *model.Errors) {
	err := api.validate.Struct(in)
	if err != nil {
		return nil, model.NewErrors(http.StatusBadRequest, err)
	}

	settings, err := api.settingsPersister.UpsertSettings(ctx, in)
	if er, ok := err.(model.Error); ok {
		if er.Code == model.ErrConcurrentUpdate {
			return nil, model.NewErrors(http.StatusConflict, er)
		}
	}
	if err != nil {
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}

	return &settings, nil
}
