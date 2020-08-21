package stdplace

import (
	"context"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	lru "github.com/hashicorp/golang-lru"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/stdtext"
)

const StdSuffix = "_std"
const topLevel = 1
const maxLevels = 4
const maxRecursion = 7
const maxRequests = 20
const timeoutMillis = 10
const cacheSize = 10000

type Standardizer struct {
	persister                 model.PlacePersister
	placeRequestChan          chan placeRequest
	placeRequestMap           map[uint32][]chan placeResponse
	placeResponseCache        *lru.TwoQueueCache
	wordRequestChan           chan wordRequest
	wordRequestMap            map[string][]chan wordResponse
	wordResponseCache         *lru.TwoQueueCache
	abbreviations             map[string]string
	typeWords                 map[string]bool
	noiseWords                map[string]bool
	largeCountries            map[uint32]bool
	mediumCountries           map[uint32]bool
	largeCountryLevelWeights  []int
	mediumCountryLevelWeights []int
	smallCountryLevelWeights  []int
	primaryMatchWeight        int
	usCountryID               uint32
}

func NewStandardizer(ctx context.Context, persister model.PlacePersister) (*Standardizer, error) {
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

	var settings *model.PlaceSettings
	for i := 0; i < 4; i++ {
		if i > 0 {
			sleepSeconds := int(math.Pow(2, float64(i-1)))
			log.Printf("[DEBUG] PlaceSettings not found; sleep for %d milliseconds\n", sleepSeconds)
			time.Sleep(time.Duration(sleepSeconds) * time.Second)
		}
		settings, err = persister.SelectPlaceSettings(ctx)
		if err == nil || !model.ErrNotFound.Matches(err) {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	std := &Standardizer{
		persister:                 persister,
		placeRequestChan:          placeRequestChan,
		placeRequestMap:           map[uint32][]chan placeResponse{},
		placeResponseCache:        placeResponseCache,
		wordRequestChan:           wordRequestChan,
		wordRequestMap:            map[string][]chan wordResponse{},
		wordResponseCache:         wordResponseCache,
		abbreviations:             settings.Abbreviations,
		typeWords:                 toStringMap(settings.TypeWords),
		noiseWords:                toStringMap(settings.NoiseWords),
		largeCountries:            toUint32Map(settings.LargeCountries),
		mediumCountries:           toUint32Map(settings.MediumCountries),
		largeCountryLevelWeights:  settings.LargeCountryLevelWeights,
		mediumCountryLevelWeights: settings.MediumCountryLevelWeights,
		smallCountryLevelWeights:  settings.SmallCountryLevelWeights,
		primaryMatchWeight:        settings.PrimaryMatchWeight,
		usCountryID:               settings.USCountryID,
	}

	go std.placeRequestListener(placeRequestChan)
	go std.wordRequestListener(wordRequestChan)

	return std, nil
}

func (ps *Standardizer) Close() {
	close(ps.placeRequestChan)
	close(ps.wordRequestChan)
}

func (ps *Standardizer) Standardize(ctx context.Context, text, defaultContainingPlace string) (*model.Place, error) {
	levelWords := tokenize(text)
	var err error
	var currentIDs []uint32
	var previousIDs []uint32
	var currentNameToken string
	lastFoundLevel := -1

	for level := len(levelWords) - 1; level >= 0; level-- {
		words := levelWords[level]
		// if all words don't match, back off and insert left-hand words as a new level (for people who don't use commas)
		wordsToSkip := 0
		var ids []uint32
		var nameType []string
		for wordsToSkip < len(words) {
			nameType = ps.getNameTypeToken(words, wordsToSkip)
			// lookup name token
			ids, err = ps.getWord(nameType[0])
			if err != nil && !model.ErrNotFound.Matches(err) {
				return nil, err
			}
			if err == nil && len(ids) > 0 {
				break
			}
			wordsToSkip++
		}
		if len(ids) > 0 && wordsToSkip > 0 {
			var newLevel []string
			for _, word := range words[0:wordsToSkip] {
				// don't push noise words ot type words down to the lower level
				if !ps.noiseWords[word] && !ps.isTypeWord(word) {
					newLevel = append(newLevel, word)
				}
			}
			if len(newLevel) > 0 {
				// insert newLevel at level
				levelWords = append(levelWords, nil)
				copy(levelWords[level+1:], levelWords[level:])
				levelWords[level] = newLevel
				level++
			}
		}

		// didn't find any matches; log and ignore
		if len(ids) == 0 {
			log.Printf("[DEBUG] Token not found text=%s word=%s\n", text, nameType[0])
		} else {
			// if we found previous matches, filter subplaces
			ignoreTypeToken := false
			if len(currentIDs) > 0 {
				matchingIDs, err := ps.filterSubplaceMatches(ids, currentIDs)
				if err != nil {
					return nil, err
				}
				// didn't find any children, try skipping over the previous level
				if len(matchingIDs) == 0 {
					skippable, err := ps.isSkippable(currentIDs)
					if err != nil {
						return nil, err
					}
					if skippable {
						// try attaching to the grandparent level if there is one
						if len(previousIDs) > 0 {
							matchingIDs, err = ps.filterSubplaceMatches(ids, previousIDs)
							if err != nil {
								return nil, err
							}
							if len(matchingIDs) > 0 {
								currentIDs = previousIDs
								log.Printf("[DEBUG] Skipping parent level text=%s\n", text)
							}
						} else {
							skippable, err = ps.isSkippable(ids)
							if err != nil {
								return nil, err
							}
							if !skippable {
								// there is no grandparent level and we matched non-skippable places, so go with what we just found
								matchingIDs = ids
								currentIDs = nil
								log.Printf("[DEBUG] Skipping parent level text=%s\n", text)
							}
						}
					}
				}

				// still didn't find any children
				if len(matchingIDs) == 0 {
					ignoreTypeToken = true // no sense matching the type if we couldn't match the name
					log.Printf("[DEBUG] Subplace matches empty text=%s word=%s\n", text, nameType[0])
					ids = currentIDs
					currentIDs = previousIDs
				} else {
					lastFoundLevel = level
					ids = matchingIDs
				}
			} else {
				// if the first match is ambiguous and we have a default containing place, filter non-top-level places outside the default containing place
				foundContainingPlace := false
				if len(ids) > 1 && defaultContainingPlace != "" {
					matchingIDs, err := ps.filterDefaultContainingPlace(ctx, ids, defaultContainingPlace)
					if err != nil {
						return nil, err
					}
					if len(matchingIDs) > 0 {
						ids = matchingIDs
						foundContainingPlace = true
					}
				}
				// if still ambiguous, require state/country match
				if len(ids) > 1 && !foundContainingPlace {
					matchingIDs, err := ps.filterTopLevel(ids)
					if err != nil {
						return nil, err
					}
					ids = matchingIDs
				}
				if len(ids) >= 0 {
					lastFoundLevel = level
				}
			}

			// if we still have multiple matches, filter on type
			if len(ids) > 1 && nameType[1] != "" && !ignoreTypeToken {
				matchingIDs, err := ps.filterTypeMatches(ids, nameType[1])
				if err != nil {
					return nil, err
				}
				// didn't find a type match
				if len(matchingIDs) == 0 {
					log.Printf("[DEBUG] Type not found text=%s type=%s\n", text, nameType[1])
				} else {
					ids = matchingIDs
				}
			}

			previousIDs = currentIDs
			currentIDs = ids
			currentNameToken = nameType[0]
		}
	}

	// if we have no matches, return not found
	if len(currentIDs) == 0 {
		log.Printf("[DEBUG] Place not found text=%s\n", text)
		return nil, model.NewError(model.ErrNotFound, text)
	}

	// remove children if we have the parents
	if len(currentIDs) > 1 {
		currentIDs, err = ps.removeChildIDs(currentIDs)
		if err != nil {
			return nil, err
		}
	}

	// if we still have multiple matches, score them and get the highest-scoring
	var place *model.Place
	if len(currentIDs) > 1 {
		bestScore := math.MinInt32
		for _, id := range currentIDs {
			p, err := ps.getPlace(id)
			if err != nil {
				return nil, err
			}
			score := ps.scoreMatch(currentNameToken, p)
			if score > bestScore {
				bestScore = score
				place = p
			}
		}
		log.Printf("[DEBUG] Ambiguous text=%s\n", text)
	} else {
		place, err = ps.getPlace(currentIDs[0])
		if err != nil {
			return nil, err
		}
	}

	// if we didn't match the last level, return "unmatched levels, best match"
	if lastFoundLevel > 0 {
		var name string
		fullName := place.FullName
		for i := lastFoundLevel - 1; i >= 0; i-- {
			name = ps.generatePlaceName(levelWords[i])
			fullName = name + ", " + fullName
		}
		place = &model.Place{
			ID:          0,
			Name:        name,
			FullName:    fullName,
			LocatedInID: place.ID,
			Level:       place.Level + lastFoundLevel,
			CountryID:   place.CountryID,
			Latitude:    place.Latitude,
			Longitude:   place.Longitude,
		}
	}

	return place, nil
}

// catenate all of the words together into one token, with ending type words in a second token
func (ps *Standardizer) getNameTypeToken(words []string, wordsToSkip int) []string {
	result := []string{"", ""}
	var tokens []string
	foundNameWord := false
	for i := len(words) - 1; i >= wordsToSkip; i-- {
		word := words[i]
		if word == "" {
			continue
		}
		// skip everything before or or now
		if i > wordsToSkip && len(tokens) > 0 && (word == "or" || word == "now") {
			break
		}
		// expand abbreviations only if there is >1 word in the phrase
		// (keeps from expanding places like No, Niigata, Japan into North)
		if len(words)-wordsToSkip > 1 {
			expansion := ps.abbreviations[word]
			if expansion != "" {
				word = expansion
			}
		}
		if !ps.typeWords[word] {
			// type words after a name word go into the type token
			if !foundNameWord && len(tokens) > 0 {
				result[1] = strings.Join(tokens, "")
				tokens = tokens[:0]
			}
			foundNameWord = true
		}
		// insert word at beginning
		tokens = append(tokens, "")
		copy(tokens[1:], tokens[0:])
		tokens[0] = word
	}
	if len(tokens) > 0 {
		result[0] = strings.Join(tokens, "")
	}
	return result
}

func (ps *Standardizer) filterDefaultContainingPlace(ctx context.Context, ids []uint32, countryText string) ([]uint32, error) {
	country, err := ps.Standardize(ctx, countryText, "")
	if err != nil {
		return nil, err
	}
	if country.ID == 0 {
		return nil, model.NewError(model.ErrNotFound, countryText)
	}
	var matchingIDs []uint32
	for _, id := range ids {
		place, err := ps.getPlace(id)
		if err != nil {
			return nil, err
		}
		// all top-level places or places in the country or places located-in the country
		// the last condition allows defaultContainingPlace to be a state or county or whatever level you want
		if place.Level == topLevel || place.CountryID == country.ID {
			matchingIDs = append(matchingIDs, id)
		} else if found, err := ps.isLocatedIn(id, country.ID, maxRecursion); found || err != nil {
			if err != nil {
				return nil, err
			}
			matchingIDs = append(matchingIDs, id)
		}
	}
	return matchingIDs, nil
}

func (ps *Standardizer) filterTypeMatches(ids []uint32, typeToken string) ([]uint32, error) {
	var matchingIDs []uint32
	for _, id := range ids {
		place, err := ps.getPlace(id)
		if err != nil {
			return nil, err
		}
		// does primary name contain the type token?
		if strings.Index(normalize(place.Name), typeToken) >= 0 {
			matchingIDs = append(matchingIDs, id)
			continue
		}
		for _, typ := range place.Types {
			// does one of the types contain the type token?
			if strings.Index(normalize(typ), typeToken) >= 0 {
				matchingIDs = append(matchingIDs, id)
				break
			}
		}
	}
	return matchingIDs, nil
}

func (ps *Standardizer) filterTopLevel(ids []uint32) ([]uint32, error) {
	var matchingIDs []uint32
	for _, id := range ids {
		place, err := ps.getPlace(id)
		if err != nil {
			return nil, err
		}
		// is the place a country or a US state?
		if place.Level == topLevel || (place.Level == topLevel+1 && place.CountryID == ps.usCountryID) {
			matchingIDs = append(matchingIDs, id)
			continue
		}
	}
	return matchingIDs, nil
}

func (ps *Standardizer) filterSubplaceMatches(ids, parentIDs []uint32) ([]uint32, error) {
	var matchingIDs []uint32
	for _, id := range ids {
		if found, err := ps.checkAncestorMatch(id, parentIDs, maxRecursion); found || err != nil {
			if err != nil {
				return nil, err
			}
			matchingIDs = append(matchingIDs, id)
		}
	}
	return matchingIDs, nil
}

func (ps *Standardizer) scoreMatch(nameToken string, place *model.Place) int {
	var weights []int
	switch {
	case ps.largeCountries[place.CountryID]:
		weights = ps.largeCountryLevelWeights
	case ps.mediumCountries[place.CountryID]:
		weights = ps.mediumCountryLevelWeights
	default:
		weights = ps.smallCountryLevelWeights
	}
	level := place.Level
	if level > maxLevels {
		level = maxLevels
	}
	score := weights[level-1]
	if strings.Index(normalize(place.Name), nameToken) >= 0 {
		score += ps.primaryMatchWeight
	}
	return score
}

func (ps *Standardizer) isTypeWord(word string) bool {
	expansion := ps.abbreviations[word]
	if expansion != "" {
		word = expansion
	}
	return ps.typeWords[word]
}

func (ps *Standardizer) generatePlaceName(words []string) string {
	last := len(words) - 1

	// ignore type words at end
	// keep cemetery as part of the full name (it's an exception; if there are others I'll create a property list)
	for last >= 0 && ps.isTypeWord(words[last]) && words[last] != "cemetery" {
		last--
	}

	// if all words are type words, keep them all
	if last < 0 {
		last = len(words) - 1
	}

	// join and capitalize
	cappedWords := make([]string, last+1)
	for i := 0; i <= last; i++ {
		cappedWords[i] = strings.Title(words[i])
	}
	return strings.Join(cappedWords, " ")
}

func (ps *Standardizer) isSkippable(ids []uint32) (bool, error) {
	for _, id := range ids {
		place, err := ps.getPlace(id)
		if err != nil {
			return false, err
		}
		if place.Level == topLevel || (place.Level == topLevel+1 && place.CountryID == ps.usCountryID) {
			return false, nil
		}
	}
	return true, nil
}

func (ps *Standardizer) removeChildIDs(ids []uint32) ([]uint32, error) {
	if len(ids) == 0 {
		return ids, nil
	}
	var result []uint32
	for _, id := range ids {
		match, err := ps.checkAncestorMatch(id, ids, maxRecursion)
		if err != nil {
			return nil, err
		}
		if !match {
			result = append(result, id)
		}
	}
	return result, nil
}

func (ps *Standardizer) isLocatedIn(id, parentID uint32, max int) (bool, error) {
	if id == parentID {
		return true, nil
	}
	place, err := ps.getPlace(id)
	if err != nil {
		return false, err
	}
	if place.LocatedInID > 0 {
		if found, err := ps.isLocatedIn(place.LocatedInID, parentID, max-1); found || err != nil {
			return found, err
		}
	}
	for _, ali := range place.AlsoLocatedInIDs {
		if found, err := ps.isLocatedIn(ali, parentID, max-1); found || err != nil {
			return found, err
		}
	}
	return false, nil
}

func (ps *Standardizer) checkAncestorMatch(id uint32, ids []uint32, max int) (bool, error) {
	place, err := ps.getPlace(id)
	if err != nil {
		return false, err
	}
	if place.LocatedInID > 0 {
		if containsUint32(ids, place.LocatedInID) {
			return true, nil
		}
		if match, err := ps.checkAncestorMatch(place.LocatedInID, ids, max-1); err != nil || match {
			return match, err
		}
	}
	for _, ali := range place.AlsoLocatedInIDs {
		if containsUint32(ids, ali) {
			return true, nil
		}
		if match, err := ps.checkAncestorMatch(ali, ids, max-1); err != nil || match {
			return match, err
		}
	}
	return false, nil
}

type placeRequest struct {
	id    uint32
	ch    chan placeResponse
	force bool
}

type wordRequest struct {
	word  string
	ch    chan wordResponse
	force bool
}

type placeResponse struct {
	place *model.Place
	err   error
}

type wordResponse struct {
	ids []uint32
	err error
}

func (ps *Standardizer) getPlace(id uint32) (*model.Place, error) {
	ch := make(chan placeResponse, 1)
	ps.placeRequestChan <- placeRequest{id: id, ch: ch}
	res := <-ch
	return res.place, res.err
}

func (ps *Standardizer) getWord(word string) ([]uint32, error) {
	ch := make(chan wordResponse, 1)
	ps.wordRequestChan <- wordRequest{word: word, ch: ch}
	res := <-ch
	return res.ids, res.err
}

func (ps *Standardizer) placeRequestListener(ch chan placeRequest) {
	ctx := context.Background()
	var cancelChan chan bool
	for req := range ch {
		if req.ch != nil {
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
		}
		// if this is the first request, set up a timeout
		if !req.force && len(ps.placeRequestMap) == 1 {
			// set up timeout
			cancelChan = make(chan bool, 1)
			go func(reqChan chan placeRequest, cancelChan chan bool) {
				select {
				case <-cancelChan:
					return
				case <-time.After(timeoutMillis * time.Millisecond):
					reqChan <- placeRequest{force: true}
				}
			}(ch, cancelChan)
		}
		if req.force || len(ps.placeRequestMap) == maxRequests {
			// cancel timeout, issue requests, and clear request map
			if cancelChan != nil {
				cancelChan <- true
				cancelChan = nil
			}
			go func(requests map[uint32][]chan placeResponse) {
				ps.issuePlaceRequests(ctx, requests)
			}(ps.placeRequestMap)
			ps.placeRequestMap = map[uint32][]chan placeResponse{}
		}
	}
}

func (ps *Standardizer) issuePlaceRequests(ctx context.Context, reqs map[uint32][]chan placeResponse) {
	if len(reqs) == 0 {
		return
	}
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
	var cancelChan chan bool
	for req := range ch {
		if req.ch != nil {
			// if in LRU cache, reply immediately
			if w, ok := ps.wordResponseCache.Get(req.word); ok {
				if ids, ok := w.([]uint32); ok {
					req.ch <- wordResponse{
						ids: ids,
						err: nil,
					}
					continue
				}
			}
			// add request to requests map
			ps.wordRequestMap[req.word] = append(ps.wordRequestMap[req.word], req.ch)
		}
		// if this is the first request, set up a timeout
		if !req.force && len(ps.wordRequestMap) == 1 {
			// set up timeout
			cancelChan = make(chan bool, 1)
			go func(reqChan chan wordRequest, cancelChan chan bool) {
				select {
				case <-cancelChan:
					return
				case <-time.After(timeoutMillis * time.Millisecond):
					reqChan <- wordRequest{force: true}
				}
			}(ch, cancelChan)
		}
		if req.force || len(ps.wordRequestMap) == maxRequests {
			// cancel timeout, issue requests, and clear request map
			if cancelChan != nil {
				cancelChan <- true
				cancelChan = nil
			}
			go func(requests map[string][]chan wordResponse) {
				ps.issueWordRequests(ctx, requests)
			}(ps.wordRequestMap)
			ps.wordRequestMap = map[string][]chan wordResponse{}
		}
	}
}

func (ps *Standardizer) issueWordRequests(ctx context.Context, reqs map[string][]chan wordResponse) {
	if len(reqs) == 0 {
		return
	}
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
		var ids []uint32
		responseErr := err
		if responseErr == nil {
			for _, placeWord := range placeWords {
				if word == placeWord.Word {
					ids = placeWord.IDs
					break
				}
			}
		}
		if responseErr == nil && len(ids) == 0 {
			responseErr = model.NewError(model.ErrNotFound, word)
		}
		for _, ch := range chs {
			ch <- wordResponse{ids: ids, err: responseErr}
		}
	}
}

func toStringMap(ss []string) map[string]bool {
	m := map[string]bool{}
	for _, s := range ss {
		m[s] = true
	}
	return m
}

func toUint32Map(is []uint32) map[uint32]bool {
	m := map[uint32]bool{}
	for _, i := range is {
		m[i] = true
	}
	return m
}

var nonAlphaNumeric = regexp.MustCompile("[^a-zA-Z0-9]+")

func tokenize(text string) [][]string {
	var levelWords [][]string

	// convert to lowercase ascii
	text = normalize(text)

	// anything after the last letter is junk
	lastPos := len(text) - 1
	for lastPos >= 0 && (text[lastPos] < 'a' || text[lastPos] > 'z') {
		lastPos--
	}
	text = text[0 : lastPos+1]

	for _, level := range strings.Split(text, ",") {
		var words []string
		for _, word := range nonAlphaNumeric.Split(level, -1) {
			if word != "" {
				words = append(words, word)
			}
		}
		if len(words) > 0 {
			levelWords = append(levelWords, words)
		}
	}

	return levelWords
}

func normalize(text string) string {
	return stdtext.AsciiFold(strings.ToLower(text))
}

func containsUint32(haystack []uint32, needle uint32) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
