package utils

import (
	"context"
	"errors"

	"github.com/ourrootsorg/cms-server/model"
)

const societyKey = "societyID"
const userKey = "userKey"
const searchUserKey = "searchUserID"

func GetSocietyIDFromContext(ctx context.Context) (uint32, error) {
	id, ok := ctx.Value(societyKey).(uint32)
	if !ok {
		return 0, errors.New("societyID not found in context")
	}
	return id, nil
}

func AddSocietyIDToContext(ctx context.Context, societyID uint32) context.Context {
	return context.WithValue(ctx, societyKey, societyID)
}

func GetUserFromContext(ctx context.Context) (*model.User, error) {
	user, ok := ctx.Value(userKey).(*model.User)
	if !ok {
		return nil, errors.New("user not found in context")
	}
	return user, nil
}

func AddUserToContext(ctx context.Context, user *model.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func GetSearchUserIDFromContext(ctx context.Context) (uint32, error) {
	id, ok := ctx.Value(searchUserKey).(uint32)
	if !ok {
		return 0, errors.New("search UserID not found in context")
	}
	return id, nil
}

func AddSearchUserIDToContext(ctx context.Context, userID uint32) context.Context {
	return context.WithValue(ctx, searchUserKey, userID)
}
