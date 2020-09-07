package dynamo

import (
	"context"
	"fmt"

	"github.com/ourrootsorg/cms-server/model"
)

func (p Persister) SelectNameVariants(ctx context.Context, nameType model.NameType, name string) (*model.NameVariants, error) {
	// TODO implement
	return nil, fmt.Errorf("SelectNameVariants not implemented")
}
