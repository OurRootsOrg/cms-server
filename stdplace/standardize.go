package stdplace

import (
	"context"
	"strconv"
	"time"

	lru "github.com/hashicorp/golang-lru"

	"github.com/ourrootsorg/cms-server/model"
)

const maxRequests = 20
const timeoutMillis = 10
const cacheSize = 1000

type Standardizer struct {
	persister          model.PlacePersister
	placeRequestChan   chan placeRequest
	placeRequestMap    map[uint32][]chan placeResponse
	placeResponseCache *lru.TwoQueueCache
	wordRequestChan    chan wordRequest
	wordRequestMap     map[string][]chan wordResponse
	wordResponseCache  *lru.TwoQueueCache
}

func NewStandardizer(persister model.PlacePersister) (*Standardizer, error) {
	placeRequestChan := make(chan placeRequest, maxRequests*2)
	wordRequestChan := make(chan wordRequest, maxRequests*2)
	placeResponseCache, err := lru.New2Q(cacheSize)
	if err != nil {
		return nil, err
	}
	wordResponseCache, err := lru.New2Q(cacheSize)
	if err != nil {
		return nil, err
	}
	std := &Standardizer{
		persister:          persister,
		placeRequestChan:   placeRequestChan,
		placeRequestMap:    map[uint32][]chan placeResponse{},
		placeResponseCache: placeResponseCache,
		wordRequestChan:    wordRequestChan,
		wordRequestMap:     map[string][]chan wordResponse{},
		wordResponseCache:  wordResponseCache,
	}
	go std.placeRequestListener(placeRequestChan)
	go std.wordRequestListener(wordRequestChan)
	return std, nil
}

type placeRequest struct {
	id uint32
	ch chan placeResponse
}

type wordRequest struct {
	word string
	ch   chan wordResponse
}

type placeResponse struct {
	place *model.Place
	err   error
}

type wordResponse struct {
	ids string
	err error
}

func (ps *Standardizer) Standardize(ctx context.Context, text string) (*model.Place, error) {
	// TODO
	return nil, nil
}

func (ps *Standardizer) Close() {
	close(ps.placeRequestChan)
	close(ps.wordRequestChan)
}

func (ps *Standardizer) getPlace(id uint32) chan placeResponse {
	ch := make(chan placeResponse, 1)
	ps.placeRequestChan <- placeRequest{id: id, ch: ch}
	return ch
}

func (ps *Standardizer) getWord(word string) chan wordResponse {
	ch := make(chan wordResponse, 1)
	ps.wordRequestChan <- wordRequest{word: word, ch: ch}
	return ch
}

func (ps *Standardizer) placeRequestListener(ch chan placeRequest) {
	ctx := context.Background()
	cancelChan := make(chan bool, 1)
	for req := range ch {
		// if in LRU cache, reply immediately
		if p, ok := ps.placeResponseCache.Get(req.id); ok {
			if place, ok := p.(model.Place); ok {
				req.ch <- placeResponse{
					place: &place,
					err:   nil,
				}
				continue
			}
		}
		// add request to requests map
		ps.placeRequestMap[req.id] = append(ps.placeRequestMap[req.id], req.ch)
		// if this is the first request, set up a timeout
		if len(ps.placeRequestMap) == 1 {
			// set up timeout
			cancelChan = make(chan bool, 1)
			go func(cancelChan chan bool) {
				select {
				case <-cancelChan:
					return
				case <-time.After(timeoutMillis * time.Millisecond):
					// issue requests and clear request map
					go func(requests map[uint32][]chan placeResponse) {
						ps.issuePlaceRequests(ctx, requests)
					}(ps.placeRequestMap)
					ps.placeRequestMap = map[uint32][]chan placeResponse{}
				}
			}(cancelChan)
		} else if len(ps.placeRequestMap) == maxRequests {
			// cancel timeout, issue requests, and clear request map
			cancelChan <- true
			go func(requests map[uint32][]chan placeResponse) {
				ps.issuePlaceRequests(ctx, requests)
			}(ps.placeRequestMap)
			ps.placeRequestMap = map[uint32][]chan placeResponse{}
		}
	}
}

func (ps *Standardizer) issuePlaceRequests(ctx context.Context, reqs map[uint32][]chan placeResponse) {
	var ids []uint32
	for id := range reqs {
		ids = append(ids, id)
	}
	places, err := ps.persister.SelectPlacesByID(ctx, ids)
	// cache place responses
	if err != nil {
		for _, place := range places {
			ps.placeResponseCache.Add(place.ID, place)
		}
	}
	// send responses
	for id, chs := range reqs {
		var responsePlace *model.Place
		responseErr := err
		if responseErr == nil {
			for _, place := range places {
				if id == place.ID {
					responsePlace = &place
					break
				}
			}
		}
		if responseErr == nil && responsePlace == nil {
			responseErr = model.NewError(model.ErrNotFound, strconv.Itoa(int(id)))
		}
		for _, ch := range chs {
			ch <- placeResponse{place: responsePlace, err: responseErr}
		}
	}
}

func (ps *Standardizer) wordRequestListener(ch chan wordRequest) {
	ctx := context.Background()
	cancelChan := make(chan bool, 1)
	for req := range ch {
		// if in LRU cache, reply immediately
		if w, ok := ps.wordResponseCache.Get(req.word); ok {
			if ids, ok := w.(string); ok {
				req.ch <- wordResponse{
					ids: ids,
					err: nil,
				}
				continue
			}
		}
		// add request to requests map
		ps.wordRequestMap[req.word] = append(ps.wordRequestMap[req.word], req.ch)
		// if this is the first request, set up a timeout
		if len(ps.wordRequestMap) == 1 {
			// set up timeout
			cancelChan = make(chan bool, 1)
			go func(cancelChan chan bool) {
				select {
				case <-cancelChan:
					return
				case <-time.After(timeoutMillis * time.Millisecond):
					// issue requests and clear request map
					go func(requests map[string][]chan wordResponse) {
						ps.issueWordRequests(ctx, requests)
					}(ps.wordRequestMap)
					ps.wordRequestMap = map[string][]chan wordResponse{}
				}
			}(cancelChan)
		} else if len(ps.wordRequestMap) == maxRequests {
			// cancel timeout, issue requests, and clear request map
			cancelChan <- true
			go func(requests map[string][]chan wordResponse) {
				ps.issueWordRequests(ctx, requests)
			}(ps.wordRequestMap)
			ps.wordRequestMap = map[string][]chan wordResponse{}
		}
	}
}

func (ps *Standardizer) issueWordRequests(ctx context.Context, reqs map[string][]chan wordResponse) {
	var words []string
	for word := range reqs {
		words = append(words, word)
	}
	placeWords, err := ps.persister.SelectPlaceWordsByWord(ctx, words)
	// cache responses
	if err != nil {
		for _, placeWord := range placeWords {
			ps.wordResponseCache.Add(placeWord.Word, placeWord.IDs)
		}
	}
	// send place responses
	for word, chs := range reqs {
		var ids string
		responseErr := err
		if responseErr == nil {
			for _, placeWord := range placeWords {
				if word == placeWord.Word {
					ids = placeWord.IDs
					break
				}
			}
		}
		if responseErr == nil && ids == "" {
			responseErr = model.NewError(model.ErrNotFound, word)
		}
		for _, ch := range chs {
			ch <- wordResponse{ids: ids, err: responseErr}
		}
	}
}
