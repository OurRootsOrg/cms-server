package api

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/ourrootsorg/cms-server/model"
	"golang.org/x/oauth2"
)

// OIDCProvider enables mocking `*oidc.Provider`
type OIDCProvider interface {
	UserInfo(ctx context.Context, tokenSource oauth2.TokenSource) (*oidc.UserInfo, error)
}

// RetrieveUser constructs or retrieves a User, either from the database or cache
func (api API) RetrieveUser(ctx context.Context, provider OIDCProvider, token *oidc.IDToken, rawToken string) (*model.User, *model.Errors) {
	var user model.User
	cacheKey := token.Issuer + "|" + token.Subject
	u, ok := api.userCache.Get(cacheKey)
	if ok {
		user, ok = u.(model.User)
	}
	if ok {
		log.Printf("[DEBUG] Found user for key '%s' in cache: %#v", cacheKey, user)
		return &user, nil
	}
	// No user in cache, so look up their info and check the database
	log.Printf("[DEBUG] No key '%s' in cache, so looking up UserInfo", cacheKey)
	oauth2Token := &oauth2.Token{
		AccessToken: rawToken,
		TokenType:   "bearer",
	}
	userInfo, err := provider.UserInfo(ctx, oauth2.StaticTokenSource(oauth2Token))
	if err != nil {
		return nil, model.NewErrors(http.StatusUnauthorized, fmt.Errorf("Failed to get userinfo: %v", err))
	}
	log.Print("[DEBUG] UserInfo:")
	log.Printf("[DEBUG] Subject: %s", userInfo.Subject)
	log.Printf("[DEBUG] Email: %s", userInfo.Email)
	log.Printf("[DEBUG] EmailVerified: %t", userInfo.EmailVerified)
	log.Printf("[DEBUG] Profile: %s", userInfo.Profile)
	userClaims := make(map[string]interface{})
	err = userInfo.Claims(&userClaims)
	if err != nil {
		log.Printf("[ERROR] Error getting claims: %v", err)
	}
	log.Printf("[DEBUG] Claims: %#v", userClaims)
	name := userClaims["name"]
	if name == nil {
		name = "<Unknown>"
	}
	ui, err := model.NewUserIn(name.(string), userInfo.Email, userInfo.EmailVerified, token.Issuer, token.Subject)
	if err != nil {
		return nil, model.NewErrors(http.StatusUnauthorized, fmt.Errorf("Failed to construct User: %v", err))
	}
	err = api.validate.Struct(ui)
	if err != nil {
		log.Printf("[ERROR] Invalid user %v", err)
		return nil, model.NewErrors(http.StatusUnauthorized, err)
	}

	// RetrieveUser will create the user if it's not already in the DB
	up, err := api.userPersister.RetrieveUser(ctx, ui)
	if err != nil {
		return nil, model.NewErrors(http.StatusUnauthorized, fmt.Errorf("Failed to retrieve user: %v", err))
	}
	// TODO: RetrieveUser doesn't update an existing user if attributes change.
	// We should probably compare `up.UserBody` to `ui.UserBody` and do an update if they're not equal.
	log.Printf("[DEBUG] Adding user '%#v' to cache with key '%s'", *up, cacheKey)
	api.userCache.Add(cacheKey, *up)
	return up, nil
}
