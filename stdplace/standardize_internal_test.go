package stdplace

import (
	"context"
	"fmt"
	"strconv"
	"sync"
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
	RequestsMutex      *sync.Mutex
	PlaceRequests      []uint32
	PlacesRequests     [][]uint32
	PlaceWordRequests  []string
	PlaceWordsRequests [][]string
}

func (pp *placePersisterMock) SelectPlaceSettings(ctx context.Context) (*model.PlaceSettings, error) {
	return &model.PlaceSettings{}, nil
}
func (pp *placePersisterMock) SelectPlace(ctx context.Context, id uint32) (*model.Place, error) {
	pp.RequestsMutex.Lock()
	pp.PlaceRequests = append(pp.PlaceRequests, id)
	pp.RequestsMutex.Unlock()
	return &model.Place{ID: id}, nil
}
func (pp *placePersisterMock) SelectPlacesByID(ctx context.Context, ids []uint32) ([]model.Place, error) {
	pp.RequestsMutex.Lock()
	pp.PlacesRequests = append(pp.PlacesRequests, ids)
	pp.RequestsMutex.Unlock()
	var result []model.Place
	for _, id := range ids {
		result = append(result, model.Place{ID: id})
	}
	return result, nil
}
func (pp *placePersisterMock) SelectPlaceWord(ctx context.Context, word string) (*model.PlaceWord, error) {
	pp.RequestsMutex.Lock()
	pp.PlaceWordRequests = append(pp.PlaceWordRequests, word)
	pp.RequestsMutex.Unlock()
	id, _ := strconv.Atoi(word)
	return &model.PlaceWord{Word: word, IDs: []uint32{uint32(id)}}, nil
}
func (pp *placePersisterMock) SelectPlaceWordsByWord(ctx context.Context, words []string) ([]model.PlaceWord, error) {
	pp.RequestsMutex.Lock()
	pp.PlaceWordsRequests = append(pp.PlaceWordsRequests, words)
	pp.RequestsMutex.Unlock()
	var result []model.PlaceWord
	for _, word := range words {
		id, _ := strconv.Atoi(word)
		result = append(result, model.PlaceWord{Word: word, IDs: []uint32{uint32(id)}})
	}
	return result, nil
}
func (pp *placePersisterMock) SelectPlacesByFullNamePrefix(ctx context.Context, prefix string, count int) ([]model.Place, error) {
	return nil, fmt.Errorf("SelectPlacesByFullNamePrefix not implemented")
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
		RequestsMutex:  &sync.Mutex{},
	}
	std, err := NewStandardizer(context.TODO(), &pp)
	assert.NoError(t, err)
	defer std.Close()

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
		RequestsMutex:      &sync.Mutex{},
	}
	std, err := NewStandardizer(context.TODO(), &pp)
	assert.NoError(t, err)
	defer std.Close()

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
		var ids []uint32
		if id, err := strconv.Atoi(rr.req); err == nil {
			ids = []uint32{uint32(id)}
		}
		assert.Equal(t, ids, rr.ids)
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
