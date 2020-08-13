package persist

import (
	"context"

	"github.com/ourrootsorg/cms-server/model"
)

const PlaceMetadataID = 1

// SelectPlaceMetadata selects the PlaceMetadata object if it exists or returns ErrNoRows
func (p PostgresPersister) SelectPlaceMetadata(ctx context.Context) (*model.PlaceMetadata, error) {
	var placeMetadata model.PlaceMetadata
	err := p.db.QueryRowContext(ctx, "SELECT body, insert_time, last_update_time FROM place_metadata WHERE id=$1", PlaceMetadataID).Scan(
		&placeMetadata.PlaceMetadataBody,
		&placeMetadata.InsertTime,
		&placeMetadata.LastUpdateTime,
	)
	id := uint32(PlaceMetadataID)
	return &placeMetadata, translateError(err, &id, nil, "")
}

// SelectPlace selects the Place object if it exists or returns ErrNoRows
func (p PostgresPersister) SelectPlace(ctx context.Context) (*model.Place, error) {
	var place model.Place
	err := p.db.QueryRowContext(ctx, "SELECT body, insert_time, last_update_time FROM place_metadata WHERE id=$1", PlaceMetadataID).Scan(
		&placeMetadata.PlaceMetadataBody,
		&placeMetadata.InsertTime,
		&placeMetadata.LastUpdateTime,
	)
	id := uint32(PlaceMetadataID)
	return &placeMetadata, translateError(err, &id, nil, "")
}
