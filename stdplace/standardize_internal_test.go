package stdplace

import (
	"context"
	"strconv"
	"testing"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/stretchr/testify/assert"
)

func TestPlaceTokenize(t *testing.T) {
	tests := []struct {
		text       string
		levelWords [][]string
	}{
		{
			text:       "Alabama",
			levelWords: [][]string{{"alabama"}},
		},
		{
			text:       "Bethel Grove, Alabama, United States",
			levelWords: [][]string{{"bethel", "grove"}, {"alabama"}, {"united", "states"}},
		},
	}

	for _, test := range tests {
		levelWords := tokenize(test.text)
		assert.Equal(t, test.levelWords, levelWords, test.text)
	}
}

type placePersisterMock struct {
	// mock.Mock
	PlaceRequests      []uint32
	PlacesRequests     [][]uint32
	PlaceWordRequests  []string
	PlaceWordsRequests [][]string
	Result             interface{}
	Errors             error
}

func (pp *placePersisterMock) SelectPlaceSettings(ctx context.Context) (*model.PlaceSettings, error) {
	return pp.Result.(*model.PlaceSettings), pp.Errors
}
func (pp *placePersisterMock) SelectPlace(ctx context.Context, id uint32) (*model.Place, error) {
	pp.PlaceRequests = append(pp.PlaceRequests, id)
	return pp.Result.(*model.Place), pp.Errors
}
func (pp *placePersisterMock) SelectPlacesByID(ctx context.Context, ids []uint32) ([]model.Place, error) {
	pp.PlacesRequests = append(pp.PlacesRequests, ids)
	var result []model.Place
	for _, id := range ids {
		result = append(result, model.Place{ID: id})
	}
	return result, pp.Errors
}
func (pp *placePersisterMock) SelectPlaceWord(ctx context.Context, word string) (*model.PlaceWord, error) {
	pp.PlaceWordRequests = append(pp.PlaceWordRequests, word)
	return pp.Result.(*model.PlaceWord), pp.Errors
}
func (pp *placePersisterMock) SelectPlaceWordsByWord(ctx context.Context, words []string) ([]model.PlaceWord, error) {
	pp.PlaceWordsRequests = append(pp.PlaceWordsRequests, words)
	var result []model.PlaceWord
	for _, word := range words {
		id, _ := strconv.Atoi(word)
		result = append(result, model.PlaceWord{Word: word, IDs: []uint32{uint32(id)}})
	}
	return result, pp.Errors
}

type placeRequestResponse struct {
	req   uint32
	place *model.Place
	err   error
}

type wordRequestResponse struct {
	req string
	ids []uint32
	err error
}

func TestGetPlace(t *testing.T) {
	pp := placePersisterMock{
		PlacesRequests: [][]uint32{},
	}
	std, err := NewStandardizer(context.TODO(), &pp)
	assert.NoError(t, err)

	// issue requests
	totalRequests := maxRequests*2 - 2
	out := make(chan placeRequestResponse, totalRequests)
	for i := 0; i < totalRequests; i++ {
		go func(id uint32) {
			place, err := std.getPlace(id)
			out <- placeRequestResponse{req: id, place: place, err: err}
		}(uint32(i))
	}
	// check responses
	m := map[uint32]bool{}
	for i := 0; i < totalRequests; i++ {
		rr := <-out
		assert.NoError(t, rr.err)
		assert.Equal(t, rr.req, rr.place.ID)
		m[rr.req] = true
	}
	assert.Equal(t, totalRequests, len(m))
	assert.Equal(t, 2, len(pp.PlacesRequests))
	total := 0
	for _, ids := range pp.PlacesRequests {
		total += len(ids)
	}
	assert.Equal(t, totalRequests, total)
}

func TestGetWord(t *testing.T) {
	pp := placePersisterMock{
		PlaceWordsRequests: [][]string{},
	}
	std, err := NewStandardizer(context.TODO(), &pp)
	assert.NoError(t, err)

	// issue requests
	totalRequests := maxRequests*2 - 2
	out := make(chan wordRequestResponse, totalRequests)
	for i := 0; i < totalRequests; i++ {
		go func(word string) {
			ids, err := std.getWord(word)
			out <- wordRequestResponse{req: word, ids: ids, err: err}
		}(strconv.Itoa(i))
	}
	// check responses
	m := map[string]bool{}
	for i := 0; i < totalRequests; i++ {
		rr := <-out
		assert.NoError(t, rr.err)
		assert.Equal(t, rr.req, rr.ids)
		m[rr.req] = true
	}
	assert.Equal(t, totalRequests, len(m))
	assert.Equal(t, 2, len(pp.PlaceWordsRequests))
	total := 0
	for _, ids := range pp.PlaceWordsRequests {
		total += len(ids)
	}
	assert.Equal(t, totalRequests, total)
}
