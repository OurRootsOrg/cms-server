package api

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/ourrootsorg/cms-server/utils"

	"github.com/ourrootsorg/cms-server/model"
)

type SocietyUserName struct {
	model.SocietyUser
	UserName string `json:"userName"`
}

func (api API) GetSocietyUserNames(ctx context.Context) ([]SocietyUserName, error) {
	// get society users
	societyUsers, err := api.societyUserPersister.SelectSocietyUsers(ctx)
	if err != nil {
		return nil, err
	}
	// get users
	var ids []uint32
	for _, societyUser := range societyUsers {
		ids = append(ids, societyUser.UserID)
	}
	users, err := api.userPersister.SelectUsersByID(ctx, ids)
	if err != nil {
		return nil, err
	}
	// join
	var societyUserNames []SocietyUserName
outer:
	for _, societyUser := range societyUsers {
		for _, user := range users {
			if user.ID == societyUser.UserID {
				societyUserNames = append(societyUserNames, SocietyUserName{
					SocietyUser: societyUser,
					UserName:    user.Name,
				})
				continue outer
			}
		}
	}
	return societyUserNames, nil
}

func (api API) UpdateSocietyUserName(ctx context.Context, id uint32, in SocietyUserName) (*SocietyUserName, error) {
	err := api.validate.Struct(in)
	if err != nil {
		return nil, NewError(err)
	}

	// can't update yourself
	user, err := utils.GetUserFromContext(ctx)
	if err != nil {
		return nil, NewError(err)
	}
	if user.ID == in.UserID {
		return nil, NewHTTPError(fmt.Errorf("cannot delete yourself"), http.StatusBadRequest)
	}
	societyUser, err := api.societyUserPersister.UpdateSocietyUser(ctx, id, in.SocietyUser)
	if err != nil {
		return nil, NewError(err)
	}
	return &SocietyUserName{
		SocietyUser: *societyUser,
		UserName:    in.UserName,
	}, nil
}

func (api API) GetSocietyUserByUser(ctx context.Context, userID uint32) (*model.SocietyUser, error) {
	var societyUser *model.SocietyUser

	// look up in cache
	cacheKey := fmt.Sprintf("%d", userID)
	u, ok := api.societyUserCache.Get(cacheKey)
	if ok {
		*societyUser, ok = u.(model.SocietyUser)
	}
	if ok {
		log.Printf("[DEBUG] Found user for key '%s' in cache: %#v", cacheKey, societyUser)
		return societyUser, nil
	}

	// read from database
	societyUser, err := api.societyUserPersister.SelectOneSocietyUserByUser(ctx, userID)
	if err != nil {
		return nil, NewError(err)
	}

	// add to cache
	api.societyUserCache.Add(cacheKey, *societyUser)

	return societyUser, nil
}

func (api API) AddSocietyUser(ctx context.Context, body model.SocietyUserBody) (*model.SocietyUser, error) {
	err := api.validate.Struct(body)
	if err != nil {
		return nil, NewError(err)
	}
	user, err := utils.GetUserFromContext(ctx)
	if err != nil {
		return nil, NewError(err)
	}
	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		return nil, NewError(err)
	}
	in := model.SocietyUserIn{
		SocietyUserBody: body,
		UserID:          user.ID,
		SocietyID:       societyID,
	}
	// add SocietyUser
	societyUser, err := api.societyUserPersister.InsertSocietyUser(ctx, in)
	if err != nil {
		return nil, NewError(err)
	}

	return societyUser, nil
}

func (api API) DeleteSocietyUser(ctx context.Context, id uint32) error {
	// can't delete yourself
	user, err := utils.GetUserFromContext(ctx)
	if err != nil {
		return NewError(err)
	}
	societyUser, err := api.societyUserPersister.SelectOneSocietyUser(ctx, id)
	if err != nil {
		return NewError(err)
	}
	if user.ID == societyUser.UserID {
		return NewHTTPError(fmt.Errorf("cannot delete yourself"), http.StatusBadRequest)
	}
	err = api.societyUserPersister.DeleteSocietyUser(ctx, id)
	if err != nil {
		return NewError(err)
	}
	return nil
}
