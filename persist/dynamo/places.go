package dynamo

import (
	"context"
	"fmt"

	"github.com/ourrootsorg/cms-server/model"
)

func (p Persister) SelectPlaceSettings(ctx context.Context) (*model.PlaceSettings, error) {
	// TODO implement
	return nil, fmt.Errorf("SelectPlaceSettings not implemented")
}
func (p Persister) SelectPlace(ctx context.Context, id uint32) (*model.Place, error) {
	// TODO implement
	return nil, fmt.Errorf("SelectPlace not implemented")
}
func (p Persister) SelectPlacesByID(ctx context.Context, ids []uint32) ([]model.Place, error) {
	// TODO implement
	return nil, fmt.Errorf("SelectPlacesByID not implemented")
}
func (p Persister) SelectPlaceWord(ctx context.Context, word string) (*model.PlaceWord, error) {
	// TODO implement
	return nil, fmt.Errorf("SelectPlaceWord not implemented")
}
func (p Persister) SelectPlaceWordsByWord(ctx context.Context, words []string) ([]model.PlaceWord, error) {
	// TODO implement
	return nil, fmt.Errorf("SelectPlaceWordsByWord not implemented")
}
func (p Persister) SelectPlacesByFullNamePrefix(ctx context.Context, prefix string, count int) ([]model.Place, error) {
	// TODO implement
	return nil, fmt.Errorf("SelectPlacesByFullNamePrefix not implemented")
}
