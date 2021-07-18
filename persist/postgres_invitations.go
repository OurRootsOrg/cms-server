package persist

import (
	"context"
	"log"

	"github.com/ourrootsorg/cms-server/utils"

	"github.com/ourrootsorg/cms-server/model"
)

// SelectInvitationsByUser selects Invitation for code
func (p PostgresPersister) SelectInvitationByCode(ctx context.Context, code string) (*model.Invitation, error) {
	var invitation model.Invitation
	log.Printf("[DEBUG] code: %s", code)
	err := p.db.QueryRowContext(ctx, "SELECT id, body, code, society_id, insert_time, last_update_time FROM invitation "+
		"WHERE code=$1", code).Scan(
		&invitation.ID,
		&invitation.InvitationBody,
		&invitation.Code,
		&invitation.SocietyID,
		&invitation.InsertTime,
		&invitation.LastUpdateTime,
	)
	if err != nil {
		return nil, translateError(err, nil, nil, "")
	}
	return &invitation, nil
}

// SelectAllInvitations selects all Invitations for society
func (p PostgresPersister) SelectInvitations(ctx context.Context) ([]model.Invitation, error) {
	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	invitations := make([]model.Invitation, 0)

	rows, err := p.db.QueryContext(ctx, "SELECT id, body, code, society_id, insert_time, last_update_time FROM invitation "+
		"WHERE society_id = $1", societyID)
	if err != nil {
		return nil, translateError(err, nil, nil, "")
	}
	defer rows.Close()
	for rows.Next() {
		var invitation model.Invitation
		err := rows.Scan(&invitation.ID, &invitation.InvitationBody, &invitation.Code, &invitation.SocietyID, &invitation.InsertTime, &invitation.LastUpdateTime)
		if err != nil {
			return nil, translateError(err, nil, nil, "")
		}
		invitations = append(invitations, invitation)
	}
	return invitations, nil
}

// SelectOneInvitation loads a single Invitation from the database
func (p PostgresPersister) SelectOneInvitation(ctx context.Context, id uint32) (*model.Invitation, error) {
	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var invitation model.Invitation
	log.Printf("[DEBUG] id: %d", id)
	err = p.db.QueryRowContext(ctx, "SELECT id, body, code, society_id, insert_time, last_update_time FROM invitation "+
		"WHERE society_id = $1 AND id=$2", societyID, id).Scan(
		&invitation.ID,
		&invitation.InvitationBody,
		&invitation.Code,
		&invitation.SocietyID,
		&invitation.InsertTime,
		&invitation.LastUpdateTime,
	)
	if err != nil {
		return nil, translateError(err, &id, nil, "")
	}
	return &invitation, nil
}

// InsertInvitation inserts a InvitationBody into the database and returns the inserted Invitation
func (p PostgresPersister) InsertInvitation(ctx context.Context, in model.InvitationIn) (*model.Invitation, error) {
	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var invitation model.Invitation
	row := p.db.QueryRowContext(ctx, "INSERT INTO invitation (body, code, society_id) VALUES ($1, $2, $3) "+
		"RETURNING id, body, code, society_id, insert_time, last_update_time", in.InvitationBody, in.Code, societyID)
	err = row.Scan(
		&invitation.ID,
		&invitation.InvitationBody,
		&invitation.Code,
		&invitation.SocietyID,
		&invitation.InsertTime,
		&invitation.LastUpdateTime,
	)
	return &invitation, translateError(err, nil, nil, "")
}

// DeleteInvitation deletes a Invitation
func (p PostgresPersister) DeleteInvitation(ctx context.Context, id uint32) error {
	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		return err
	}
	_, err = p.db.ExecContext(ctx, "DELETE FROM invitation WHERE society_id = $1 AND id = $2", societyID, id)
	return translateError(err, &id, nil, "")
}
