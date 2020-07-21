package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/elastic/go-elasticsearch/v7/esutil"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/ourrootsorg/cms-server/model"
)

// given and surname fuzziness constants are flags that can be OR'd together
// names can also contain wildcards (*, ?) but in that case fuzziness is ignored
const (
	FuzzyNameAlternate        = 1 << iota // 1 - alternate spellings - not yet implemented
	FuzzyNameSoundsLikeNarrow = 1 << iota // 2 - sounds-like (narrow) - high-precision, low-recall
	FuzzyNameSoundsLikeBroad  = 1 << iota // 4 - sounds-like (broad) - low-precision, high-recall
	FuzzyNameLevenshtein      = 1 << iota // 8 - fuzzy (levenshtein)
	FuzzyNameInitials         = 1 << iota // 16 - initials (applies only to given)
)

// place fuzziness constants can also be OR'd together
const (
	FuzzyPlaceHigherJurisdictions = 1 << iota // 1 - searches for City, County, State, Country also match County, State, Country or State, Country - not yet implemented
	FuzzyPlaceNearby              = 1 << iota // 2 - not yet implemented
)

// date and place searches are not yet implemented

// date fuzziness is simply a +/- number of years to generate a year range
// e.g., a birthDate of 1880 and a birthDateFuzziness of 5 would result in a year range 1875..1885

// SearchRequest contains the possible search request parameters
type SearchRequest struct {
	// name
	Given            string `schema:"given"`
	GivenFuzziness   int    `schema:"givenFuzziness"`
	Surname          string `schema:"surname"`
	SurnameFuzziness int    `schema:"surnameFuzziness"`
	// relatives
	FatherGiven            string `schema:"fatherGiven"`
	FatherGivenFuzziness   int    `schema:"fatherGivenFuzziness"`
	FatherSurname          string `schema:"fatherSurname"`
	FatherSurnameFuzziness int    `schema:"fatherSurnameFuzziness"`
	MotherGiven            string `schema:"motherGiven"`
	MotherGivenFuzziness   int    `schema:"motherGivenFuzziness"`
	MotherSurname          string `schema:"motherSurname"`
	MotherSurnameFuzziness int    `schema:"motherSurnameFuzziness"`
	SpouseGiven            string `schema:"spouseGiven"`
	SpouseGivenFuzziness   int    `schema:"spouseGivenFuzziness"`
	SpouseSurname          string `schema:"spouseSurname"`
	SpouseSurnameFuzziness int    `schema:"spouseSurnameFuzziness"`
	OtherGiven             string `schema:"otherGiven"`
	OtherGivenFuzziness    int    `schema:"otherGivenFuzziness"`
	OtherSurname           string `schema:"otherSurname"`
	OtherSurnameFuzziness  int    `schema:"otherSurnameFuzziness"`
	// events
	BirthDate               string `schema:"birthDate"`
	BirthDateFuzziness      int    `schema:"birthDateFuzziness"`
	BirthPlace              string `schema:"birthPlace"`
	BirthPlaceFuzziness     int    `schema:"birthPlaceFuzziness"`
	MarriageDate            string `schema:"marriageDate"`
	MarriageDateFuzziness   int    `schema:"marriageDateFuzziness"`
	MarriagePlace           string `schema:"marriagePlace"`
	MarriagePlaceFuzziness  int    `schema:"marriagePlaceFuzziness"`
	ResidenceDate           string `schema:"residenceDate"`
	ResidenceDateFuzziness  int    `schema:"residenceDateFuzziness"`
	ResidencePlace          string `schema:"residencePlace"`
	ResidencePlaceFuzziness int    `schema:"residencePlaceFuzziness"`
	DeathDate               string `schema:"deathDate"`
	DeathDateFuzziness      int    `schema:"deathDateFuzziness"`
	DeathPlace              string `schema:"deathPlace"`
	DeathPlaceFuzziness     int    `schema:"deathPlaceFuzziness"`
	AnyDate                 string `schema:"anyDate"` // match on any date
	AnyDateFuzziness        int    `schema:"anyDateFuzziness"`
	AnyPlace                string `schema:"anyPlace"` // match on any place
	AnyPlaceFuzziness       int    `schema:"anyPlaceFuzziness"`
	// other
	Keywords string `schema:"keywords"`
	// event facets and filters
	BirthPlace1Facet      bool   `schema:"birthPlace1Facet"`
	BirthPlace1           string `schema:"birthPlace1"`
	BirthPlace2Facet      bool   `schema:"birthPlace2Facet"`
	BirthPlace2           string `schema:"birthPlace2"`
	BirthPlace3Facet      bool   `schema:"birthPlace3Facet"`
	BirthPlace3           string `schema:"birthPlace3"`
	BirthCenturyFacet     bool   `schema:"birthCenturyFacet"`
	BirthCentury          string `schema:"birthCentury"`
	BirthDecadeFacet      bool   `schema:"birthDecadeFacet"`
	BirthDecade           string `schema:"birthDecade"`
	MarriagePlace1Facet   bool   `schema:"marriagePlace1Facet"`
	MarriagePlace1        string `schema:"marriagePlace1"`
	MarriagePlace2Facet   bool   `schema:"marriagePlace2Facet"`
	MarriagePlace2        string `schema:"marriagePlace2"`
	MarriagePlace3Facet   bool   `schema:"marriagePlace3Facet"`
	MarriagePlace3        string `schema:"marriagePlace3"`
	MarriageCenturyFacet  bool   `schema:"marriageCenturyFacet"`
	MarriageCentury       string `schema:"marriageCentury"`
	MarriageDecadeFacet   bool   `schema:"marriageDecadeFacet"`
	MarriageDecade        string `schema:"marriageDecade"`
	ResidencePlace1Facet  bool   `schema:"residencePlace1Facet"`
	ResidencePlace1       string `schema:"residencePlace1"`
	ResidencePlace2Facet  bool   `schema:"residencePlace2Facet"`
	ResidencePlace2       string `schema:"residencePlace2"`
	ResidencePlace3Facet  bool   `schema:"residencePlace3Facet"`
	ResidencePlace3       string `schema:"residencePlace3"`
	ResidenceCenturyFacet bool   `schema:"residenceCenturyFacet"`
	ResidenceCentury      string `schema:"residenceCentury"`
	ResidenceDecadeFacet  bool   `schema:"residenceDecadeFacet"`
	ResidenceDecade       string `schema:"residenceDecade"`
	DeathPlace1Facet      bool   `schema:"deathPlace1Facet"`
	DeathPlace1           string `schema:"deathPlace1"`
	DeathPlace2Facet      bool   `schema:"deathPlace2Facet"`
	DeathPlace2           string `schema:"deathPlace2"`
	DeathPlace3Facet      bool   `schema:"deathPlace3Facet"`
	DeathPlace3           string `schema:"deathPlace3"`
	DeathCenturyFacet     bool   `schema:"deathCenturyFacet"`
	DeathCentury          string `schema:"deathCentury"`
	DeathDecadeFacet      bool   `schema:"deathDecadeFacet"`
	DeathDecade           string `schema:"deathDecade"`
	OtherPlace1Facet      bool   `schema:"otherPlace1Facet"`
	OtherPlace1           string `schema:"otherPlace1"`
	OtherPlace2Facet      bool   `schema:"otherPlace2Facet"`
	OtherPlace2           string `schema:"otherPlace2"`
	OtherPlace3Facet      bool   `schema:"otherPlace3Facet"`
	OtherPlace3           string `schema:"otherPlace3"`
	OtherCenturyFacet     bool   `schema:"otherCenturyFacet"`
	OtherCentury          string `schema:"otherCentury"`
	OtherDecadeFacet      bool   `schema:"otherDecadeFacet"`
	OtherDecade           string `schema:"otherDecade"`
	// other facets and filters
	CategoryFacet   bool   `schema:"categoryFacet"`
	Category        string `schema:"category"`
	CollectionFacet bool   `schema:"collectionFacet"`
	Collection      string `schema:"collection"`
}

// int
type Search struct {
	Query  Query          `json:"query,omitempty"`
	Aggs   map[string]Agg `json:"aggs,omitempty"`
	Source []string       `json:"_source,omitempty"`
}
type Query struct {
	Bool     *BoolQuery            `json:"bool,omitempty"`
	DisMax   *DisMaxQuery          `json:"dis_max,omitempty"`
	Fuzzy    map[string]FuzzyQuery `json:"fuzzy,omitempty"`
	Match    map[string]MatchQuery `json:"match,omitempty"`
	Range    map[string]RangeQuery `json:"range,omitempty"`
	Term     map[string]TermQuery  `json:"term,omitempty"`
	Wildcard map[string]TermQuery  `json:"wildcard,omitempty"`
}
type BoolQuery struct {
	Must   []Query `json:"must,omitempty"`
	Should []Query `json:"should,omitempty"`
	Filter []Query `json:"filter,omitempty"`
}
type DisMaxQuery struct {
	Queries []Query `json:"queries,omitempty"`
}
type FuzzyQuery struct {
	Value     string  `json:"value"`
	Fuzziness string  `json:"fuzziness,omitempty"`
	Rewrite   string  `json:"rewrite,omitempty"`
	Boost     float32 `json:"boost,omitempty"`
}
type MatchQuery struct {
	Query string  `json:"query"`
	Boost float32 `json:"boost,omitempty"`
}
type RangeQuery struct {
	GTE   int     `json:"gte,omitempty"`
	LTE   int     `json:"lte,omitempty"`
	Boost float32 `json:"boost,omitempty"`
}
type TermQuery struct {
	Value string  `json:"value"`
	Boost float32 `json:"boost,omitempty"`
}
type Agg struct {
	Terms *TermsAgg `json:"terms,omitempty"`
	Range *RangeAgg `json:"range,omitempty"`
}
type TermsAgg struct {
	Field string `json:"field"`
	Size  int    `json:"size,omitempty"`
}
type RangeAgg struct {
	Field  string          `json:"field"`
	Keyed  bool            `json:"keyed,omitempty"`
	Ranges []RangeAggRange `json:"ranges"`
}
type RangeAggRange struct {
	Key  string `json:"key"`
	From int    `json:"from,omitempty'"`
	To   int    `json:"to,omitempty'"`
}

type ESErrorResponse struct {
	Error ESError `json:"error"`
}
type ESError struct {
	Type   string `json:"string"`
	Reason string `json:"reason"`
}

type ESSearchResponse struct {
	Took     int            `json:"took"`
	TimedOut bool           `json:"timed_out"`
	Shards   ESSearchShards `json:"_shards"`
	Hits     ESSearchHits   `json:"hits"`
}
type ESSearchShards struct {
	Total      int `json:"total"`
	Successful int `json:"successful"`
	Skipped    int `json:"skipped"`
	Failed     int `json:"failed"`
}
type ESSearchHits struct {
	Total    ESSearchTotal `json:"total"`
	MaxScore float64       `json:"max_score"`
	Hits     []ESSearchHit `json:"hits"`
}
type ESSearchTotal struct {
	Value    int    `json:"value"`
	Relation string `json:"relation"`
}
type ESSearchHit struct {
	ID      string         `json:"_id"`
	Version int            `json:"_version"` // only in search by id
	Found   bool           `json:"found"`    // only in search by id
	Score   float64        `json:"_score"`   // only in search
	Source  ESSearchSource `json:"_source"`
}
type ESSearchSource struct {
	CollectionID uint32 `json:"collectionId"`
}

type HitData struct {
	ID           string
	RecordID     uint32
	Role         string
	CollectionID uint32
}

const numWorkers = 5

type GivenSurname struct {
	given   string
	surname string
}

type nameExtractor func(GivenSurname) string

var IndexRoles = map[string]string{
	"principal":   "",
	"father":      "f",
	"mother":      "m",
	"spouse":      "s",
	"bride":       "b",
	"groom":       "g",
	"brideFather": "bf",
	"brideMother": "bm",
	"groomFather": "gf",
	"groomMother": "gm",
	"other":       "o",
}

var IndexRolesReversed = reverseMap(IndexRoles)

func reverseMap(m map[string]string) map[string]string {
	result := map[string]string{}
	for k, v := range m {
		result[v] = k
	}
	return result
}

var Relatives = []string{"father", "mother", "spouse", "other"}

var RelativeRoles = map[string]map[string][]string{
	"principal": {
		"father": {"father"},
		"mother": {"mother"},
		"spouse": {"spouse"},
		"other":  {"bride", "groom", "brideFather", "brideMother", "groomFather", "groomMother", "other"},
	},
	"father": {
		"father": {},
		"mother": {},
		"spouse": {"mother"},
		"other":  {"principal", "spouse", "bride", "groom", "brideFather", "brideMother", "groomFather", "groomMother", "other"},
	},
	"mother": {
		"father": {},
		"mother": {},
		"spouse": {"father"},
		"other":  {"principal", "spouse", "bride", "groom", "brideFather", "brideMother", "groomFather", "groomMother", "other"},
	},
	"spouse": {
		"father": {},
		"mother": {},
		"spouse": {"principal"},
		"other":  {"father", "mother", "bride", "groom", "brideFather", "brideMother", "groomFather", "groomMother", "other"},
	},
	"bride": {
		"father": {"brideFather"},
		"mother": {"brideMother"},
		"spouse": {"groom"},
		"other":  {"principal", "father", "mother", "spouse", "groomFather", "groomMother", "other"},
	},
	"groom": {
		"father": {"groomFather"},
		"mother": {"groomMother"},
		"spouse": {"bride"},
		"other":  {"principal", "father", "mother", "spouse", "brideFather", "brideMother", "other"},
	},
	"brideFather": {
		"father": {},
		"mother": {},
		"spouse": {"brideMother"},
		"other":  {"principal", "father", "mother", "spouse", "bride", "groom", "groomFather", "groomMother", "other"},
	},
	"brideMother": {
		"father": {},
		"mother": {},
		"spouse": {"brideFather"},
		"other":  {"principal", "father", "mother", "spouse", "bride", "groom", "groomFather", "groomMother", "other"},
	},
	"groomFather": {
		"father": {},
		"mother": {},
		"spouse": {"groomMother"},
		"other":  {"principal", "father", "mother", "spouse", "bride", "groom", "brideFather", "brideMother", "other"},
	},
	"groomMother": {
		"father": {},
		"mother": {},
		"spouse": {"groomFather"},
		"other":  {"principal", "father", "mother", "spouse", "bride", "groom", "brideFather", "brideMother", "other"},
	},
	"other": {
		"father": {},
		"mother": {},
		"spouse": {},
		"other":  {"principal", "father", "mother", "spouse", "bride", "groom", "brideFather", "brideMother", "groomFather", "groomMother"},
	},
}

// IndexPost
func (api API) IndexPost(ctx context.Context, post *model.Post) error {
	var countSuccessful uint64

	lastModified := strconv.FormatInt(time.Now().Unix()*1000, 10)

	// read collection for post
	collection, errs := api.GetCollection(ctx, post.Collection)
	if errs != nil {
		log.Printf("[ERROR] GetCollection %v\n", errs)
		return errs
	}
	// read collection for post
	categories, errs := api.GetCategoriesByID(ctx, collection.Categories)
	if errs != nil {
		log.Printf("[ERROR] GetCategory %v\n", errs)
		return errs
	}
	// read records for post
	records, errs := api.GetRecordsForPost(ctx, post.ID)
	if errs != nil {
		log.Printf("[ERROR] GetRecordsForPost %v\n", errs)
		return errs
	}

	// create the bulk indexer
	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:      "records",  // The default index name
		Client:     api.es,     // The Elasticsearch client
		NumWorkers: numWorkers, // The number of worker goroutines
	})
	if err != nil {
		log.Printf("[ERROR] Error creating the bulk indexer: %s", err)
		return err
	}
	biClosed := false
	defer func() {
		if !biClosed {
			_ = bi.Close(ctx)
		}
	}()

	for _, record := range records.Records {
		err = indexRecord(&record, post, collection, categories, lastModified, &countSuccessful, bi)
		if err != nil {
			log.Printf("[ERROR] Unexpected error %d: %v", record.ID, err)
			return err
		}
	}

	if err := bi.Close(ctx); err != nil {
		log.Printf("[ERROR] Unexpected error %v\n", err)
		return err
	}
	biClosed = true

	biStats := bi.Stats()
	if biStats.NumFailed > 0 {
		msg := fmt.Sprintf("[ERROR] Failed to index %d records\n", biStats.NumFailed)
		log.Printf(msg)
		return errors.New(msg)
	}

	log.Printf("[INFO] Indexed %d records\n", biStats.NumFlushed)
	return nil
}

func indexRecord(record *model.Record, post *model.Post, collection *model.Collection, categories []model.Category,
	lastModified string, countSuccessful *uint64, bi esutil.BulkIndexer) error {

	for role, suffix := range IndexRoles {
		if suffix != "" {
			suffix = "_" + suffix
		}
		// get data for role
		data := getDataForRole(collection.Mappings, record, role)
		//log.Printf("!!! role=%s record=%v data=%v\n", role, record, data)
		_, givenFound := data["given"]
		_, surnameFound := data["surname"]
		if !givenFound && !surnameFound {
			continue
		}

		// get relatives' names
		for _, relative := range Relatives {
			names := getNames(collection.Mappings, record, RelativeRoles[role][relative])
			givens := getNameParts(names, func(name GivenSurname) string { return name.given })
			surnames := getNameParts(names, func(name GivenSurname) string { return name.surname })
			if len(givens) > 0 {
				data[relative+"Given"] = strings.Join(givens, " ")
			}
			if len(surnames) > 0 {
				data[relative+"Surname"] = strings.Join(surnames, " ")
			}
		}

		// get other data
		var catNames []string
		for _, cat := range categories {
			catNames = append(catNames, cat.Name)
		}
		data["post"] = post.ID
		data["collection"] = collection.Name
		data["collectionId"] = collection.ID
		data["category"] = catNames
		data["lastModified"] = lastModified

		// add to BulkIndexer
		bs, err := json.Marshal(data)
		if err != nil {
			log.Printf("[ERROR] encoding record %d: %v", record.ID, err)
			return err
		}

		// Add an item to the BulkIndexer
		err = bi.Add(
			context.Background(),
			esutil.BulkIndexerItem{
				// Action field configures the operation to perform (index, create, delete, update)
				Action: "index",

				DocumentID: strconv.Itoa(int(record.ID)) + suffix,

				// Body is an `io.Reader` with the payload
				Body: bytes.NewReader(bs),

				// OnSuccess is called for each successful operation
				OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
					atomic.AddUint64(countSuccessful, 1)
				},

				// OnFailure is called for each failed operation
				OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
					if err != nil {
						log.Printf("[ERROR]: %s", err)
					} else {
						log.Printf("[ERROR]: %s: %s", res.Error.Type, res.Error.Reason)
					}
				},
			},
		)
		if err != nil {
			log.Printf("[ERROR] indexing record %d: %v\n", record.ID, err)
		}
	}

	return nil
}

func (api API) SearchByID(ctx context.Context, id string) (*model.SearchHit, error) {
	res, err := api.es.Get("records", id,
		api.es.Get.WithContext(ctx),
	)
	if err != nil {
		log.Printf("[ERROR] SearchByID %v", err)
		return nil, NewError(err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e ESErrorResponse
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Printf("Error parsing the response body: %v", err)
			return nil, NewError(err)
		} else {
			// Print the response status and error information.
			msg := fmt.Sprintf("[%s] %s: %s id=%s", res.Status(), e.Error.Type, e.Error.Reason, id)
			log.Println(msg)
			if res.StatusCode == http.StatusNotFound {
				return nil, NewError(model.NewError(model.ErrNotFound, id))
			}
			return nil, NewError(errors.New(msg))
		}
	}

	// get hit data
	var hit ESSearchHit
	if err := json.NewDecoder(res.Body).Decode(&hit); err != nil {
		log.Printf("Error parsing the response body: %s\n", err)
		return nil, NewError(err)
	}
	if !hit.Found {
		log.Printf("[ERROR] record ID %s not found\n", id)
		return nil, NewError(model.NewError(model.ErrNotFound, id))
	}
	hitData, err := getHitData(hit)
	if err != nil {
		log.Printf("[ERROR] getting hit data %v\n", err)
		return nil, NewError(err)
	}

	// read record and collection
	record, errs := api.GetRecord(ctx, hitData.RecordID)
	if errs != nil {
		log.Printf("[ERROR] record not found %d\n", hitData.RecordID)
		return nil, errs
	}
	collection, errs := api.GetCollection(ctx, hitData.CollectionID)
	if errs != nil {
		log.Printf("[ERROR] collection not found %d\n", hitData.CollectionID)
		return nil, errs
	}

	return &model.SearchHit{
		ID:             hitData.ID,
		Person:         constructSearchPerson(collection.Mappings, hitData.Role, record),
		Record:         constructSearchRecord(collection.Mappings, record),
		CollectionName: collection.Name,
		CollectionID:   collection.ID,
	}, nil
}

func (api API) SearchDeleteByPost(ctx context.Context, id uint32) error {
	search := Search{
		Query: Query{
			Term: map[string]TermQuery{
				"post": {
					Value: strconv.Itoa(int(id)),
				},
			},
		},
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(search); err != nil {
		log.Printf("[ERROR] encoding delete by post query %v\n", err)
		return NewError(err)
	}
	res, err := api.es.DeleteByQuery([]string{"records"}, &buf,
		api.es.DeleteByQuery.WithContext(ctx),
	)
	if err != nil {
		log.Printf("[ERROR] SearchDeleteByPost %v", err)
		return NewError(err)
	}
	defer res.Body.Close()
	return nil
}

func (api API) SearchDeleteByID(ctx context.Context, id string) error {
	res, err := api.es.Delete("records", id,
		api.es.Delete.WithContext(ctx),
	)
	if err != nil {
		log.Printf("[ERROR] DeleteByID %v", err)
		return NewError(err)
	}
	defer res.Body.Close()
	return nil
}

// Search
func (api API) Search(ctx context.Context, req *SearchRequest) (*model.SearchResult, error) {
	search := constructSearchQuery(req)

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(search); err != nil {
		log.Printf("[ERROR] encoding query %v\n", err)
		return nil, NewError(err)
	}
	log.Printf("[DEBUG] Request=%v Query=%s\n", req, string(buf.Bytes()))

	res, err := api.es.Search(
		api.es.Search.WithContext(ctx),
		api.es.Search.WithIndex("records"),
		api.es.Search.WithBody(&buf),
		api.es.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		log.Printf("[ERROR] Search %v", err)
		return nil, NewError(err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e ESErrorResponse
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Printf("Error parsing the response body: %v", err)
			return nil, NewError(err)
		} else {
			// Print the response status and error information.
			msg := fmt.Sprintf("[%s] %s: %s", res.Status(), e.Error.Type, e.Error.Reason)
			log.Println(msg)
			return nil, NewError(errors.New(msg))
		}
	}

	// get hit datas
	var r ESSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Printf("Error parsing the response body: %s\n", err)
		return nil, NewError(err)
	}
	var hitDatas []HitData
	var recordIDs []uint32
	var collectionIDs []uint32
	for _, hit := range r.Hits.Hits {
		hitData, err := getHitData(hit)
		if err != nil {
			return nil, NewError(err)
		}
		hitDatas = append(hitDatas, *hitData)
		recordIDs = append(recordIDs, hitData.RecordID)
		found := false
		for _, id := range collectionIDs {
			if hitData.CollectionID == id {
				found = true
				break
			}
		}
		if !found {
			collectionIDs = append(collectionIDs, hitData.CollectionID)
		}
	}

	// read records and collections
	records, errs := api.GetRecordsByID(ctx, recordIDs)
	if errs != nil {
		return nil, errs
	}
	collections, errs := api.GetCollectionsByID(ctx, collectionIDs)
	if errs != nil {
		return nil, errs
	}

	// construct search hits
	hits := []model.SearchHit{}
	for _, hitData := range hitDatas {
		// get record
		var record model.Record
		found := false
		for _, item := range records {
			if item.ID == hitData.RecordID {
				found = true
				record = item
				break
			}
		}
		if !found {
			msg := fmt.Sprintf("[ERROR] record %d not found\n", hitData.RecordID)
			return nil, NewError(errors.New(msg))
		}

		// get collection
		var collection model.Collection
		found = false
		for _, item := range collections {
			if item.ID == hitData.CollectionID {
				found = true
				collection = item
				break
			}
		}
		if !found {
			msg := fmt.Sprintf("[ERROR] collection %d not found\n", hitData.CollectionID)
			return nil, NewError(errors.New(msg))
		}

		// construct search hit
		hits = append(hits, model.SearchHit{
			ID:     hitData.ID,
			Person: constructSearchPerson(collection.Mappings, hitData.Role, &record),
			//Record:         constructSearchRecord(&record), // only return record in search by id
			CollectionName: collection.Name,
			CollectionID:   collection.ID,
		})
	}

	return &model.SearchResult{
		Total:    r.Hits.Total.Value,
		MaxScore: r.Hits.MaxScore,
		Hits:     hits,
	}, nil
}

func constructSearchQuery(req *SearchRequest) *Search {
	mustQueries := []Query{}
	shouldQueries := []Query{}
	filterQueries := []Query{}
	aggs := map[string]Agg{}

	// name
	givenQueries := constructNameQueries("given", req.Given, req.GivenFuzziness, true)
	surnameQueries := constructNameQueries("surname", req.Surname, req.SurnameFuzziness, false)
	if len(givenQueries) > 0 || len(surnameQueries) > 0 {
		mustQueries = append(mustQueries, Query{
			Bool: &BoolQuery{
				Should: append(givenQueries, surnameQueries...),
			},
		})
	}

	// relative names
	shouldQueries = append(shouldQueries, constructNameQueries("fatherGiven", req.FatherGiven, req.FatherGivenFuzziness, true)...)
	shouldQueries = append(shouldQueries, constructNameQueries("fatherSurname", req.FatherSurname, req.FatherSurnameFuzziness, false)...)
	shouldQueries = append(shouldQueries, constructNameQueries("motherGiven", req.MotherGiven, req.MotherGivenFuzziness, true)...)
	shouldQueries = append(shouldQueries, constructNameQueries("motherSurname", req.MotherSurname, req.MotherSurnameFuzziness, false)...)
	shouldQueries = append(shouldQueries, constructNameQueries("spouseGiven", req.SpouseGiven, req.SpouseGivenFuzziness, true)...)
	shouldQueries = append(shouldQueries, constructNameQueries("spouseSurname", req.SpouseSurname, req.SpouseSurnameFuzziness, false)...)
	shouldQueries = append(shouldQueries, constructNameQueries("otherGiven", req.OtherGiven, req.OtherGivenFuzziness, true)...)
	shouldQueries = append(shouldQueries, constructNameQueries("otherSurname", req.OtherSurname, req.OtherSurnameFuzziness, false)...)

	// events
	shouldQueries = append(shouldQueries, constructDateQueries("birthYear", "birthDate", req.BirthDate, req.BirthDateFuzziness)...)
	shouldQueries = append(shouldQueries, constructPlaceQueries("birthPlace", req.BirthPlace, req.BirthPlaceFuzziness)...)
	shouldQueries = append(shouldQueries, constructDateQueries("marriageYear", "marriageDate", req.MarriageDate, req.MarriageDateFuzziness)...)
	shouldQueries = append(shouldQueries, constructPlaceQueries("marriagePlace", req.MarriagePlace, req.MarriagePlaceFuzziness)...)
	shouldQueries = append(shouldQueries, constructDateQueries("residenceYear", "residenceDate", req.ResidenceDate, req.ResidenceDateFuzziness)...)
	shouldQueries = append(shouldQueries, constructPlaceQueries("residencePlace", req.ResidencePlace, req.ResidencePlaceFuzziness)...)
	shouldQueries = append(shouldQueries, constructDateQueries("deathYear", "deathDate", req.DeathDate, req.DeathDateFuzziness)...)
	shouldQueries = append(shouldQueries, constructPlaceQueries("deathPlace", req.DeathPlace, req.DeathPlaceFuzziness)...)

	// any place
	if len(req.AnyPlace) > 0 {
		var anyPlaceQueries []Query
		anyPlaceQueries = append(anyPlaceQueries, constructPlaceQueries("birthPlace", req.AnyPlace, req.AnyPlaceFuzziness)...)
		anyPlaceQueries = append(anyPlaceQueries, constructPlaceQueries("marriagePlace", req.AnyPlace, req.AnyPlaceFuzziness)...)
		anyPlaceQueries = append(anyPlaceQueries, constructPlaceQueries("residencePlace", req.AnyPlace, req.AnyPlaceFuzziness)...)
		anyPlaceQueries = append(anyPlaceQueries, constructPlaceQueries("deathPlace", req.AnyPlace, req.AnyPlaceFuzziness)...)
		anyPlaceQueries = append(anyPlaceQueries, constructPlaceQueries("otherPlace", req.AnyPlace, req.AnyPlaceFuzziness)...)
		shouldQueries = append(shouldQueries, Query{
			DisMax: &DisMaxQuery{
				Queries: anyPlaceQueries,
			},
		})
	}

	// other
	shouldQueries = append(shouldQueries, constructTextQueries("keywords", req.Keywords)...)

	// filters
	filterQueries = append(filterQueries, constructFilterQueries("category", req.Category)...)
	filterQueries = append(filterQueries, constructFilterQueries("collection", req.Collection)...)
	filterQueries = append(filterQueries, constructFilterQueries("birthPlace1", req.BirthPlace1)...)
	filterQueries = append(filterQueries, constructFilterQueries("birthPlace2", req.BirthPlace2)...)
	filterQueries = append(filterQueries, constructFilterQueries("birthPlace3", req.BirthPlace3)...)
	filterQueries = append(filterQueries, constructCenturyFilterQueries("birthDecade", req.BirthCentury)...)
	filterQueries = append(filterQueries, constructFilterQueries("birthDecade", req.BirthDecade)...)
	filterQueries = append(filterQueries, constructFilterQueries("marriagePlace1", req.MarriagePlace1)...)
	filterQueries = append(filterQueries, constructFilterQueries("marriagePlace2", req.MarriagePlace2)...)
	filterQueries = append(filterQueries, constructFilterQueries("marriagePlace3", req.MarriagePlace3)...)
	filterQueries = append(filterQueries, constructCenturyFilterQueries("marriageDecade", req.MarriageCentury)...)
	filterQueries = append(filterQueries, constructFilterQueries("marriageDecade", req.MarriageDecade)...)
	filterQueries = append(filterQueries, constructFilterQueries("residencePlace1", req.ResidencePlace1)...)
	filterQueries = append(filterQueries, constructFilterQueries("residencePlace2", req.ResidencePlace2)...)
	filterQueries = append(filterQueries, constructFilterQueries("residencePlace3", req.ResidencePlace3)...)
	filterQueries = append(filterQueries, constructCenturyFilterQueries("residenceDecade", req.ResidenceCentury)...)
	filterQueries = append(filterQueries, constructFilterQueries("residenceDecade", req.ResidenceDecade)...)
	filterQueries = append(filterQueries, constructFilterQueries("deathPlace1", req.DeathPlace1)...)
	filterQueries = append(filterQueries, constructFilterQueries("deathPlace2", req.DeathPlace2)...)
	filterQueries = append(filterQueries, constructFilterQueries("deathPlace3", req.DeathPlace3)...)
	filterQueries = append(filterQueries, constructCenturyFilterQueries("deathDecade", req.DeathCentury)...)
	filterQueries = append(filterQueries, constructFilterQueries("deathDecade", req.DeathDecade)...)
	filterQueries = append(filterQueries, constructFilterQueries("otherPlace1", req.OtherPlace1)...)
	filterQueries = append(filterQueries, constructFilterQueries("otherPlace2", req.OtherPlace2)...)
	filterQueries = append(filterQueries, constructFilterQueries("otherPlace3", req.OtherPlace3)...)
	filterQueries = append(filterQueries, constructCenturyFilterQueries("otherDecade", req.OtherCentury)...)
	filterQueries = append(filterQueries, constructFilterQueries("otherDecade", req.OtherDecade)...)

	// facets
	addTermsAgg(aggs, "category", req.CategoryFacet)
	addTermsAgg(aggs, "collection", len(req.Category) > 0 && req.CollectionFacet)
	addTermsAgg(aggs, "birthPlace1", req.BirthPlace1Facet)
	addTermsAgg(aggs, "birthPlace2", len(req.BirthPlace1) > 0 && req.BirthPlace2Facet)
	addTermsAgg(aggs, "birthPlace3", len(req.BirthPlace1) > 0 && len(req.BirthPlace2) > 0 && req.BirthPlace3Facet)
	addCenturyAgg(aggs, "birthCentury", "birthDecade", req.BirthCenturyFacet)
	addTermsAgg(aggs, "birthDecade", len(req.BirthCentury) > 0 && req.BirthDecadeFacet)
	addTermsAgg(aggs, "marriagePlace1", req.MarriagePlace1Facet)
	addTermsAgg(aggs, "marriagePlace2", len(req.MarriagePlace1) > 0 && req.MarriagePlace2Facet)
	addTermsAgg(aggs, "marriagePlace3", len(req.MarriagePlace1) > 0 && len(req.MarriagePlace2) > 0 && req.MarriagePlace3Facet)
	addCenturyAgg(aggs, "marriageCentury", "marriageDecade", req.MarriageCenturyFacet)
	addTermsAgg(aggs, "marriageDecade", len(req.MarriageCentury) > 0 && req.MarriageDecadeFacet)
	addTermsAgg(aggs, "residencePlace1", req.ResidencePlace1Facet)
	addTermsAgg(aggs, "residencePlace2", len(req.ResidencePlace1) > 0 && req.ResidencePlace2Facet)
	addTermsAgg(aggs, "residencePlace3", len(req.ResidencePlace1) > 0 && len(req.ResidencePlace2) > 0 && req.ResidencePlace3Facet)
	addCenturyAgg(aggs, "residenceCentury", "residenceDecade", req.ResidenceCenturyFacet)
	addTermsAgg(aggs, "residenceDecade", len(req.ResidenceCentury) > 0 && req.ResidenceDecadeFacet)
	addTermsAgg(aggs, "deathPlace1", req.DeathPlace1Facet)
	addTermsAgg(aggs, "deathPlace2", len(req.DeathPlace1) > 0 && req.DeathPlace2Facet)
	addTermsAgg(aggs, "deathPlace3", len(req.DeathPlace1) > 0 && len(req.DeathPlace2) > 0 && req.DeathPlace3Facet)
	addCenturyAgg(aggs, "deathCentury", "deathDecade", req.DeathCenturyFacet)
	addTermsAgg(aggs, "deathDecade", len(req.DeathCentury) > 0 && req.DeathDecadeFacet)
	addTermsAgg(aggs, "otherPlace1", req.OtherPlace1Facet)
	addTermsAgg(aggs, "otherPlace2", len(req.OtherPlace1) > 0 && req.OtherPlace2Facet)
	addTermsAgg(aggs, "otherPlace3", len(req.OtherPlace1) > 0 && len(req.OtherPlace2) > 0 && req.OtherPlace3Facet)
	addCenturyAgg(aggs, "otherCentury", "otherDecade", req.OtherCenturyFacet)
	addTermsAgg(aggs, "otherDecade", len(req.OtherCentury) > 0 && req.OtherDecadeFacet)

	return &Search{
		Query: Query{
			Bool: &BoolQuery{
				Must:   mustQueries,
				Should: shouldQueries,
				Filter: filterQueries,
			},
		},
		Aggs: aggs,
	}
}

// TODO learn the best boost values
const exactNameBoost = 1.0
const narrowNameBoost = 0.8
const wildcardNameBoost = 0.7
const broadNameBoost = 0.6
const initialNameBoost = 0.4
const fuzzyNameBoost = 0.2

func constructNameQueries(label, value string, fuzziness int, isGiven bool) []Query {
	if len(value) == 0 {
		return nil
	}
	var queries []Query

	for _, v := range splitWord(value) {
		if strings.ContainsAny(v, "*?") {
			v, err := asciifold(strings.ToLower(v))
			if err != nil {
				log.Printf("[INFO] unable to fold %s\n", v)
				v = strings.ToLower(v)
			}

			// TODO disallow wildcards within the first 3 characters?
			if strings.HasPrefix(v, "*") || strings.HasPrefix(v, "?") {
				continue
			}
			queries = append(queries, Query{
				Wildcard: map[string]TermQuery{
					label: {
						Value: v,
						Boost: wildcardNameBoost,
					},
				},
			})
			continue
		}

		exactQuery := Query{
			Match: map[string]MatchQuery{
				label: {
					Query: v,
					Boost: exactNameBoost,
				},
			},
		}

		if fuzziness == 0 {
			queries = append(queries, exactQuery)
			continue
		}

		subqueries := []Query{exactQuery}

		if fuzziness&FuzzyNameAlternate > 0 {
			// TODO alternate spellings
		}
		// TODO choose the best coders for broad and narrow
		if fuzziness&FuzzyNameSoundsLikeNarrow > 0 {
			subqueries = append(subqueries, Query{
				Match: map[string]MatchQuery{
					label + ".narrow": {
						Query: v,
						Boost: narrowNameBoost,
					},
				},
			})
		}
		if fuzziness&FuzzyNameSoundsLikeBroad > 0 {
			subqueries = append(subqueries, Query{
				Match: map[string]MatchQuery{
					label + ".broad": {
						Query: v,
						Boost: broadNameBoost,
					},
				},
			})
		}
		if fuzziness&FuzzyNameLevenshtein > 0 {
			std, err := asciifold(strings.ToLower(v))
			if err != nil {
				log.Printf("[INFO] unable to fold %s\n", v)
				std = strings.ToLower(v)
			}
			subqueries = append(subqueries, Query{
				Fuzzy: map[string]FuzzyQuery{
					label: {
						Value:     std,
						Fuzziness: "AUTO",
						Rewrite:   "constant_score_boolean",
						Boost:     fuzzyNameBoost,
					},
				},
			})
		}
		if fuzziness&FuzzyNameInitials > 0 && isGiven {
			subqueries = append(subqueries, Query{
				Match: map[string]MatchQuery{
					label: {
						Query: v[0:1],
						Boost: initialNameBoost,
					},
				},
			})
		}

		queries = append(queries, Query{
			DisMax: &DisMaxQuery{
				Queries: subqueries,
			},
		})
	}
	return queries
}

const exactYearBoost = 0.7
const rangeYearBoost = 0.3

func constructDateQueries(yearLabel, dateLabel, value string, fuzziness int) []Query {
	if len(value) < 4 {
		return nil
	}
	var q Query

	// just accept years for now
	if len(value) > 4 {
		value = value[0:4]
	}
	year, err := strconv.Atoi(value)

	if err != nil || fuzziness <= 0 || fuzziness > 10 {
		q = Query{
			Term: map[string]TermQuery{
				yearLabel: {
					Value: value,
					Boost: exactYearBoost,
				},
			},
		}
	} else {
		q = Query{
			Range: map[string]RangeQuery{
				yearLabel: {
					GTE:   year - fuzziness,
					LTE:   year + fuzziness,
					Boost: rangeYearBoost,
				},
			},
		}
	}
	return []Query{q}
}

const exactPlaceBoost = 1.0
const wildcardPlaceBoost = 0.7
const fuzzyPlaceBoost = 0.2
const levelPlaceBoost = 0.2

func constructPlaceQueries(label, value string, fuzziness int) []Query {
	if len(value) == 0 {
		return nil
	}

	if strings.ContainsAny(value, "~*?") {
		var queries []Query
		for _, v := range splitWord(value) {
			v, err := asciifold(strings.ToLower(v))
			if err != nil {
				log.Printf("[INFO] unable to fold %s\n", v)
			}
			if strings.HasPrefix(v, "~") && !strings.ContainsAny(v, "*?") {
				queries = append(queries, Query{
					Fuzzy: map[string]FuzzyQuery{
						label: {
							Value:     v[1:],
							Fuzziness: "AUTO",
							Rewrite:   "constant_score_boolean",
							Boost:     fuzzyPlaceBoost,
						},
					},
				})
				continue
			}
			v = strings.ReplaceAll(v, "~", "")

			if strings.ContainsAny(v, "*?") {
				// TODO disallow wildcards within the first 3 characters?
				if strings.HasPrefix(value, "*") || strings.HasPrefix(value, "?") {
					continue
				}
				queries = append(queries, Query{
					Wildcard: map[string]TermQuery{
						label: {
							Value: v,
							Boost: wildcardPlaceBoost,
						},
					},
				})
				continue
			}

			queries = append(queries, Query{
				Term: map[string]TermQuery{
					label: {
						Value: v,
						Boost: exactPlaceBoost,
					},
				},
			})
		}

		return queries
	}

	levels := splitPlace(value)
	reverse(levels)
	// limit to 4 levels
	if len(levels) > 4 {
		levels = levels[0:4]
	}

	queries := []Query{
		{
			Term: map[string]TermQuery{
				fmt.Sprintf("%s%d", label, len(levels)): {
					Value: strings.Join(levels, "|"),
					Boost: exactPlaceBoost,
				},
			},
		},
	}

	if fuzziness&FuzzyPlaceHigherJurisdictions > 0 {
		for i := 1; i < len(levels); i++ {
			// don't match on just "United States"
			if i == 1 && levels[0] == "United States" {
				continue
			}
			queries = append(queries, Query{
				Term: map[string]TermQuery{
					fmt.Sprintf("%s%d", label, i+1): {
						Value: strings.Join(levels[0:i], "|") + "|_",
						Boost: float32(i) * levelPlaceBoost,
					},
				},
			})
		}
	}

	// TODO include nearby places (lat and lon)
	if fuzziness&FuzzyPlaceNearby > 0 {

	}

	return queries
}

func constructTextQueries(label, value string) []Query {
	if len(value) == 0 {
		return nil
	}
	return []Query{
		{
			Match: map[string]MatchQuery{
				label: {
					Query: value,
				},
			},
		},
	}
}

func constructFilterQueries(label, value string) []Query {
	if len(value) == 0 {
		return nil
	}
	return []Query{
		{
			Term: map[string]TermQuery{
				label: {
					Value: value,
				},
			},
		},
	}
}

func constructCenturyFilterQueries(label, value string) []Query {
	if len(value) == 0 {
		return nil
	}
	gte := 0
	lte := 0
	if strings.HasPrefix(value, "<") {
		lte = 1599
	} else {
		if val, err := strconv.Atoi(value); err == nil {
			gte = val
			lte = val + 99
		}
	}
	return []Query{
		{
			Range: map[string]RangeQuery{
				label: {
					GTE: gte,
					LTE: lte,
				},
			},
		},
	}
}

func addTermsAgg(aggs map[string]Agg, label string, cond bool) {
	if cond {
		aggs[label] = Agg{
			Terms: &TermsAgg{
				Field: label,
				Size:  250,
			},
		}
	}
}

func addCenturyAgg(aggs map[string]Agg, label, field string, cond bool) {
	if cond {
		aggs[label] = Agg{
			Range: &RangeAgg{
				Field: field,
				Keyed: true,
				Ranges: []RangeAggRange{
					{
						Key: "<1600",
						To:  1600,
					},
					{
						Key:  "1600",
						From: 1600,
						To:   1700,
					},
					{
						Key:  "1700",
						From: 1700,
						To:   1800,
					},
					{
						Key:  "1800",
						From: 1800,
						To:   1900,
					},
					{
						Key:  "1900",
						From: 1900,
						To:   2000,
					},
					{
						Key:  "2000",
						From: 2000,
					},
				},
			},
		}
	}
}

func splitWord(name string) []string {
	re := regexp.MustCompile("[^\\pL*?~]+") // keep ~*? for fuzzy and wildcards
	return re.Split(name, -1)
}

func splitPlace(place string) []string {
	re := regexp.MustCompile("\\s*,\\s*")
	return re.Split(place, -1)
}

func reverse(s []string) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// try to match lucene's asciifolding
var transliterations = map[rune]string{
	'Æ': "AE",
	'Ð': "D",
	'Ł': "L",
	'Ø': "O",
	'Þ': "Th",
	'ß': "ss",
	'ẞ': "SS",
	'æ': "ae",
	'ð': "d",
	'ł': "l",
	'ø': "o",
	'þ': "th",
	'Œ': "OE",
	'œ': "oe",
}

type Transliterator struct {
}

func (t Transliterator) Transform(dst, src []byte, atEOF bool) (int, int, error) {
	var err error
	total := 0
	for i, w := 0, 0; i < len(src) && err == nil; i += w {
		var n int
		r, width := utf8.DecodeRune(src[i:])
		if d, ok := transliterations[r]; ok {
			n = copy(dst[total:], d)
			if n < len(d) {
				err = transform.ErrShortDst
			}
		} else {
			n = copy(dst[total:], src[i:i+width])
			if n < width {
				err = transform.ErrShortDst
			}
		}
		total += n
		w = width
	}

	return total, len(src), err
}

func (t Transliterator) Reset() {
}

func asciifold(s string) (string, error) {
	var tl Transliterator
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), tl, norm.NFC)
	result, _, err := transform.String(t, s)
	if err != nil {
		result = s // return as-is
	}
	return result, err
}

func getHitData(r ESSearchHit) (*HitData, error) {
	idParts := strings.Split(r.ID, "_")
	role := "principal"
	if len(idParts) > 1 {
		role = IndexRolesReversed[idParts[1]]
	}
	rid, err := strconv.Atoi(idParts[0])
	if err != nil {
		return nil, err
	}
	if r.Source.CollectionID == 0 {
		msg := fmt.Sprintf("Missing collectionID for ID %s\n", r.ID)
		log.Printf("[ERROR] %s\n", msg)
		return nil, errors.New(msg)
	}

	return &HitData{
		ID:           r.ID,
		RecordID:     uint32(rid),
		Role:         role,
		CollectionID: r.Source.CollectionID,
	}, nil
}

func constructSearchPerson(mappings []model.CollectionMapping, role string, record *model.Record) model.SearchPerson {
	data := getDataForRole(mappings, record, role)
	// TODO populate events
	events := []model.SearchEvent{}
	relationships := []model.SearchRelationship{}
	for _, relative := range Relatives {
		names := getNames(mappings, record, RelativeRoles[role][relative])
		if len(names) > 0 {
			relationships = append(relationships, model.SearchRelationship{
				Type: relative,
				Name: strings.Join(getNameParts(names, func(name GivenSurname) string { return fmt.Sprintf("%s %s", name.given, name.surname) }), ", "),
			})
		}
	}
	return model.SearchPerson{
		Name:          fmt.Sprintf("%s %s", data["given"], data["surname"]),
		Role:          role,
		Events:        events,
		Relationships: relationships,
	}
}

func constructSearchRecord(mappings []model.CollectionMapping, record *model.Record) model.SearchRecord {
	lvs := []model.SearchLabelValue{}
	for _, mapping := range mappings {
		if mapping.DbField == "" || record.Data[mapping.Header] == "" {
			continue
		}
		lvs = append(lvs, model.SearchLabelValue{
			Label: mapping.DbField,
			Value: record.Data[mapping.Header],
		})
	}
	return lvs
}

func getDataForRole(mappings []model.CollectionMapping, record *model.Record, role string) map[string]interface{} {
	data := map[string]interface{}{}

	// TODO get marriage data for spouse too
	for _, mapping := range mappings {
		if mapping.IxRole == role && record.Data[mapping.Header] != "" {
			data[mapping.IxField] = record.Data[mapping.Header]
		}
	}
	return data
}

func getNames(mappings []model.CollectionMapping, record *model.Record, roles []string) []GivenSurname {
	names := []GivenSurname{}

	for _, role := range roles {
		var givens []string
		var surnames []string
		for _, mapping := range mappings {
			if mapping.IxRole == role {
				if mapping.IxField == "given" && record.Data[mapping.Header] != "" {
					givens = append(givens, record.Data[mapping.Header])
				} else if mapping.IxField == "surname" {
					surnames = append(surnames, record.Data[mapping.Header])
				}
			}
		}
		if len(givens) > 0 || len(surnames) > 0 {
			names = append(names, GivenSurname{
				given:   strings.Join(givens, " "),
				surname: strings.Join(surnames, " "),
			})
		}
	}
	return names
}

func getNameParts(names []GivenSurname, extractor nameExtractor) []string {
	var parts []string
	for _, name := range names {
		part := extractor(name)
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}
