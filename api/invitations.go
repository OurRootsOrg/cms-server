package api

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/utils"
)

type InvitationSocietyName struct {
	model.Invitation
	SocietyName string `json:"societyName"`
}

func (api API) GetInvitations(ctx context.Context) ([]model.Invitation, error) {
	invitations, err := api.invitationPersister.SelectInvitations(ctx)
	if err != nil {
		return nil, NewError(err)
	}
	return invitations, nil
}

func (api API) AddInvitation(ctx context.Context, body model.InvitationBody) (*model.Invitation, error) {
	err := api.validate.Struct(body)
	if err != nil {
		return nil, NewError(err)
	}
	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		return nil, NewError(err)
	}

	// generate code
	b := make([]byte, 8)
	_, err = rand.Read(b)
	if err != nil {
		return nil, NewError(err)
	}
	code := fmt.Sprintf("%d-%x", societyID, b)

	in := model.InvitationIn{
		InvitationBody: body,
		SocietyID:      societyID,
		Code:           code,
	}

	invitation, err := api.invitationPersister.InsertInvitation(ctx, in)
	if err != nil {
		return nil, NewError(err)
	}
	return invitation, nil
}

func (api API) DeleteInvitation(ctx context.Context, id uint32) error {
	err := api.invitationPersister.DeleteInvitation(ctx, id)
	if err != nil {
		return NewError(err)
	}
	return nil
}

func (api API) GetInvitationSocietyName(ctx context.Context, code string) (*InvitationSocietyName, error) {
	// get invitation
	invitation, err := api.invitationPersister.SelectInvitationByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	// get society
	society, err := api.societyPersister.SelectSocietySummary(ctx, invitation.SocietyID)
	if err != nil {
		return nil, err
	}
	return &InvitationSocietyName{
		Invitation:  *invitation,
		SocietyName: society.Name,
	}, nil
}

func (api API) AcceptInvitation(ctx context.Context, code string) (*model.SocietyUser, error) {
	// read user
	user, err := utils.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	// read invitation
	invitation, err := api.invitationPersister.SelectInvitationByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	// set society context
	sctx := utils.AddSocietyIDToContext(ctx, invitation.SocietyID)
	// add or update society user
	societyUser, err := api.societyUserPersister.SelectOneSocietyUserByUser(sctx, user.ID)
	if err == nil {
		societyUser.Level = model.AuthLevel(math.Max(float64(societyUser.Level), float64(invitation.Level)))
		societyUser, err = api.societyUserPersister.UpdateSocietyUser(sctx, societyUser.ID, *societyUser)
	} else {
		societyUserIn := model.SocietyUserIn{
			SocietyUserBody: model.SocietyUserBody{
				Level: invitation.Level,
			},
			UserID:    user.ID,
			SocietyID: invitation.SocietyID,
		}
		societyUser, err = api.societyUserPersister.InsertSocietyUser(sctx, societyUserIn)
	}
	if err != nil {
		return nil, err
	}
	// delete invitation
	err = api.invitationPersister.DeleteInvitation(sctx, invitation.ID)
	if err != nil {
		return nil, err
	}
	return societyUser, nil
}
