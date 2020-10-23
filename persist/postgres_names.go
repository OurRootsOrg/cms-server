package persist

import (
	"context"
	"fmt"

	"github.com/ourrootsorg/cms-server/model"
)

// SelectNameVariants selects the NameVariants object if it exists or returns ErrNoRows
func (p PostgresPersister) SelectNameVariants(ctx context.Context, nameType model.NameType, name string) (*model.NameVariants, error) {
	var nameVariants model.NameVariants
	var table string
	switch nameType {
	case model.GivenType:
		table = "givenname_variants"
	case model.SurnameType:
		table = "surname_variants"
	default:
		return nil, model.NewError(model.ErrOther, fmt.Sprintf("Unknown name type %d", nameType))
	}

	err := p.db.QueryRowContext(ctx, "SELECT name, variants, insert_time, last_update_time FROM "+table+" WHERE name = $1", name).Scan(
		&nameVariants.Name,
		&nameVariants.Variants,
		&nameVariants.InsertTime,
		&nameVariants.LastUpdateTime,
	)
	return &nameVariants, translateError(err, nil, nil, "")
}
