package stdplace

import (
	"context"
	"strconv"
	"testing"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/stretchr/testify/assert"
)

type placePersisterMock struct {
	// mock.Mock
	PlaceRequests      []uint32
	PlacesRequests     [][]uint32
	PlaceWordRequests  []string
	PlaceWordsRequests [][]string
	Result             interface{}
	Errors             error
}

func (pp *placePersisterMock) SelectPlaceMetadata(ctx context.Context) (*model.PlaceMetadata, error) {
	return pp.Result.(*model.PlaceMetadata), pp.Errors
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
		result = append(result, model.PlaceWord{Word: word, IDs: word})
	}
	return result, pp.Errors
}

type placeRequestResponse struct {
	req uint32
	res placeResponse
}

type wordRequestResponse struct {
	req string
	res wordResponse
}

func TestGetPlace(t *testing.T) {
	pp := placePersisterMock{
		PlacesRequests: [][]uint32{},
	}
	std, err := NewStandardizer(&pp)
	assert.NoError(t, err)

	// issue requests
	totalRequests := maxRequests*2 - 2
	out := make(chan placeRequestResponse, totalRequests)
	for i := 0; i < totalRequests; i++ {
		go func(id uint32) {
			res := <-std.getPlace(id)
			out <- placeRequestResponse{req: id, res: res}
		}(uint32(i))
	}
	// check responses
	m := map[uint32]bool{}
	for i := 0; i < totalRequests; i++ {
		rr := <-out
		assert.NoError(t, rr.res.err)
		assert.Equal(t, rr.req, rr.res.place.ID)
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
	std, err := NewStandardizer(&pp)
	assert.NoError(t, err)

	// issue requests
	totalRequests := maxRequests*2 - 2
	out := make(chan wordRequestResponse, totalRequests)
	for i := 0; i < totalRequests; i++ {
		go func(word string) {
			res := <-std.getWord(word)
			out <- wordRequestResponse{req: word, res: res}
		}(strconv.Itoa(i))
	}
	// check responses
	m := map[string]bool{}
	for i := 0; i < totalRequests; i++ {
		rr := <-out
		assert.NoError(t, rr.res.err)
		assert.Equal(t, rr.req, rr.res.ids)
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
