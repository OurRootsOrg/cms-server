package api

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/utils"
)

func (api API) GetSocietySummariesForCurrentUser(ctx context.Context) ([]model.SocietySummary, error) {
	user, err := utils.GetUserFromContext(ctx)
	if err != nil {
		return nil, NewError(err)
	}
	societyUsers, err := api.societyUserPersister.SelectAllSocietyUsersByUser(ctx, user.ID)
	if err != nil {
		return nil, NewError(err)
	}
	var ids []uint32
	for _, societyUser := range societyUsers {
		ids = append(ids, societyUser.SocietyID)
	}
	societySummaries, err := api.societyPersister.SelectSocietySummariesByID(ctx, ids)
	if err != nil {
		return nil, NewError(err)
	}
	return societySummaries, nil
}

func (api API) GetSocietySummary(ctx context.Context) (*model.SocietySummary, error) {
	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	societySummary, err := api.societyPersister.SelectSocietySummary(ctx, societyID)
	if err != nil {
		return nil, NewError(err)
	}
	return societySummary, nil
}

func (api API) GetSociety(ctx context.Context) (*model.Society, error) {
	society, err := api.societyPersister.SelectSociety(ctx)
	if err != nil {
		return nil, NewError(err)
	}
	return society, nil
}

func (api API) AddSociety(ctx context.Context, in model.SocietyIn) (*model.Society, error) {
	err := api.validate.Struct(in)
	if err != nil {
		return nil, NewError(err)
	}
	user, err := utils.GetUserFromContext(ctx)
	if err != nil {
		return nil, NewError(err)
	}

	// add secret key
	if in.SecretKey == "" {
		b := make([]byte, 16)
		_, err = rand.Read(b)
		if err != nil {
			return nil, NewError(err)
		}
		in.SecretKey = fmt.Sprintf("%x", b)
	}
	// add society
	society, err := api.societyPersister.InsertSociety(ctx, in)
	if err != nil {
		return nil, NewError(err)
	}
	sctx := utils.AddSocietyIDToContext(ctx, society.ID)

	// add user to society
	societyUser := model.SocietyUserIn{
		SocietyUserBody: model.SocietyUserBody{
			Level: model.AuthAdmin,
		},
		UserID:    user.ID,
		SocietyID: society.ID,
	}
	_, err = api.societyUserPersister.InsertSocietyUser(sctx, societyUser)
	if err != nil {
		return nil, NewError(err)
	}

	return society, nil
}

func (api API) UpdateSociety(ctx context.Context, in model.Society) (*model.Society, error) {
	err := api.validate.Struct(in)
	if err != nil {
		return nil, NewError(err)
	}
	society, err := api.societyPersister.UpdateSociety(ctx, in)
	if err != nil {
		return nil, NewError(err)
	}
	return society, nil
}

func (api API) DeleteSociety(ctx context.Context) error {
	// TODO lots of things to delete here
	err := api.societyPersister.DeleteSociety(ctx)
	if err != nil {
		return NewError(err)
	}
	return nil
}
