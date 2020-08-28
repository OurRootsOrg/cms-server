package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ourrootsorg/cms-server/stddate"
	"github.com/ourrootsorg/cms-server/stdplace"

	"github.com/ourrootsorg/cms-server/stdtext"

	"github.com/elastic/go-elasticsearch/v7/esutil"

	"github.com/ourrootsorg/cms-server/model"
)

// given and surname fuzziness constants are flags that can be OR'd together
// names can also contain wildcards (*, ?) but in that case fuzziness above Exact is ignored
const FuzzyNameDefault = 0
const (
	FuzzyNameExact            = 1 << iota // 1 - exact match
	FuzzyNameAlternate        = 1 << iota // 2 - alternate spellings - not yet implemented
	FuzzyNameSoundsLikeNarrow = 1 << iota // 4 - sounds-like (narrow) - high-precision, low-recall
	FuzzyNameSoundsLikeBroad  = 1 << iota // 8 - sounds-like (broad) - low-precision, high-recall
	FuzzyNameLevenshtein      = 1 << iota // 16 - fuzzy (levenshtein)
	FuzzyNameInitials         = 1 << iota // 32 - initials (applies only to given)
)

// date fuzziness constants cannot by OR'd together
const (
	FuzzyDateDefault = iota // 0 - default
	FuzzyDateExact   = iota // 1 - exact year
	FuzzyDateOne     = iota // 2 - +/- 1 year
	FuzzyDateTwo     = iota // 3 - +/- 2 years
	FuzzyDateFive    = iota // 4 - +/- 5 years
	FuzzyDateTen     = iota // 5 - +/- 10 years
)

// place fuzziness constants can also be OR'd together
const FuzzyPlaceDefault = 0
const (
	FuzzyPlaceExact               = 1 << iota // 1 - search this place only
	FuzzyPlaceHigherJurisdictions = 1 << iota // 2 - searches for City, County, State, Country also match County, State, Country or State, Country
	FuzzyPlaceNearby              = 1 << iota // 4 - search nearby places // TODO not implemented
)

const MaxFrom = 1000
const MaxSize = 100
const DefaultSize = 10

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
	// facets and filters
	CollectionPlace1Facet bool   `schema:"collectionPlace1Facet"`
	CollectionPlace1      string `schema:"collectionPlace1"`
	CollectionPlace2Facet bool   `schema:"collectionPlace2Facet"`
	CollectionPlace2      string `schema:"collectionPlace2"`
	CollectionPlace3Facet bool   `schema:"collectionPlace3Facet"`
	CollectionPlace3      string `schema:"collectionPlace3"`
	CategoryFacet         bool   `schema:"categoryFacet"`
	Category              string `schema:"category"`
	CollectionFacet       bool   `schema:"collectionFacet"`
	Collection            string `schema:"collection"`
	// from and size
	From int `schema:"from"`
	Size int `schema:"size"`
}

// int
type Search struct {
	Query  Query          `json:"query,omitempty"`
	Aggs   map[string]Agg `json:"aggs,omitempty"`
	Source []string       `json:"_source,omitempty"`
	From   int            `json:"from,omitempty"`
	Size   int            `json:"size,omitempty"`
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
	Took         int                            `json:"took"`
	TimedOut     bool                           `json:"timed_out"`
	Shards       ESSearchShards                 `json:"_shards"`
	Hits         ESSearchHits                   `json:"hits"`
	Aggregations map[string]ESSearchAggregation `json:"aggregations"`
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
type ESSearchAggregation struct {
	DocCountErrorUpperBound int                         `json:"doc_count_error_upper_bound"`
	SumOtherDocCount        int                         `json:"sum_other_doc_count"`
	Buckets                 []ESSearchAggregationBucket `json:"buckets"`
}
type ESSearchAggregationBucket struct {
	Key      string `json:"key"`
	DocCount int    `json:"doc_count"`
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

var EventTypes = []string{"birth", "marriage", "residence", "death", "other"}

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
	// read categories for post
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

		// populate the record to index
		ixRecord := map[string]interface{}{}
		ixRecord["given"] = data["given"]
		ixRecord["surname"] = data["surname"]
		if ixRecord["given"] == "" && ixRecord["surname"] == "" {
			log.Printf("[DEBUG] No given name or surname found for record %#v, mappings %#v, role %s",
				record, collection.Mappings, role)
			continue
		}

		// get relatives' names
		for _, relative := range Relatives {
			names := getNames(collection.Mappings, record, RelativeRoles[role][relative])
			givens := getNameParts(names, func(name GivenSurname) string { return name.given })
			surnames := getNameParts(names, func(name GivenSurname) string { return name.surname })
			if len(givens) > 0 {
				ixRecord[relative+"Given"] = strings.Join(givens, " ")
			}
			if len(surnames) > 0 {
				ixRecord[relative+"Surname"] = strings.Join(surnames, " ")
			}
		}

		// get events
		for _, eventType := range EventTypes {
			if data[eventType+"Date"] != "" {
				dates, years, valid := getDatesYears(data[eventType+"Date_std"])
				if valid {
					ixRecord[eventType+"DateStd"] = dates
					ixRecord[eventType+"Year"] = years
				}
			}
			if data[eventType+"Place"] != "" {
				placeLevels := getPlaceLevels(data[eventType+"Place_std"])
				if len(placeLevels) > 0 {
					ixRecord[eventType+"Place"] = data[eventType+"Place"]
					ixRecord[eventType+"Place1"] = placeLevels[0]
				}
				if len(placeLevels) > 1 {
					ixRecord[eventType+"Place2"] = placeLevels[1]
				}
				if len(placeLevels) > 2 {
					ixRecord[eventType+"Place3"] = placeLevels[2]
				}
				if len(placeLevels) > 3 {
					ixRecord[eventType+"Place4"] = placeLevels[3]
				}
			}
		}

		// keywords
		ixRecord["keywords"] = data["keywords"]

		// get other data
		var catNames []string
		for _, cat := range categories {
			catNames = append(catNames, cat.Name)
		}
		ixRecord["post"] = post.ID
		ixRecord["category"] = catNames
		ixRecord["collection"] = collection.Name
		ixRecord["collectionId"] = collection.ID
		if collection.Location != "" {
			placeLevels := getPlaceFacets(collection.Location)
			if len(placeLevels) > 0 {
				ixRecord["collectionPlace1"] = placeLevels[0]
			}
			if len(placeLevels) > 1 {
				ixRecord["collectionPlace2"] = placeLevels[1]
			}
			if len(placeLevels) > 2 {
				ixRecord["collectionPlace3"] = placeLevels[2]
			}
		}
		ixRecord["lastModified"] = lastModified

		// add to BulkIndexer
		bs, err := json.Marshal(ixRecord)
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
			log.Printf("[DEBUG] ES error: %s", msg)
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
		ID:                 hitData.ID,
		Person:             constructSearchPerson(collection.Mappings, hitData.Role, record),
		Record:             constructSearchRecord(collection.Mappings, record),
		CollectionName:     collection.Name,
		CollectionID:       collection.ID,
		CollectionLocation: collection.Location,
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

	var buf2 bytes.Buffer
	tee := io.TeeReader(res.Body, &buf2)
	s, _ := ioutil.ReadAll(tee)
	log.Printf("[DEBUG] %s\n", s)
	r2 := bytes.NewReader(buf2.Bytes())

	// get hit datas
	var r ESSearchResponse
	if err := json.NewDecoder(r2).Decode(&r); err != nil {
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
			//Location:       collection.Location, // only return location in search by id
		})
	}

	// construct facets
	facets := map[string]model.SearchFacet{}
	for key, aggr := range r.Aggregations {
		var buckets []model.SearchFacetBucket
		for _, bucket := range aggr.Buckets {
			buckets = append(buckets, model.SearchFacetBucket{
				Label: bucket.Key,
				Count: bucket.DocCount,
			})
		}
		facets[key] = model.SearchFacet{
			ErrorUpperBound: aggr.DocCountErrorUpperBound,
			OtherDocCount:   aggr.SumOtherDocCount,
			Buckets:         buckets,
		}
	}

	return &model.SearchResult{
		Total:    r.Hits.Total.Value,
		MaxScore: r.Hits.MaxScore,
		Hits:     hits,
		Facets:   facets,
	}, nil
}

func constructSearchQuery(req *SearchRequest) *Search {
	var mustQueries []Query
	var shouldQueries []Query
	var filterQueries []Query
	aggs := map[string]Agg{}

	// name
	shouldGivenQueries, mustGivenQueries := constructNameQueries("given", req.Given, req.GivenFuzziness, true)
	shouldSurnameQueries, mustSurnameQueries := constructNameQueries("surname", req.Surname, req.SurnameFuzziness, false)
	if len(shouldGivenQueries) > 0 || len(shouldSurnameQueries) > 0 || len(mustGivenQueries) > 0 || len(mustSurnameQueries) > 0 {
		mustQueries = append(mustQueries, Query{
			Bool: &BoolQuery{
				Must:   append(mustGivenQueries, mustSurnameQueries...),
				Should: append(shouldGivenQueries, shouldSurnameQueries...),
			},
		})
	}

	// relative names
	shouldSubqueries, mustSubqueries := constructNameQueries("fatherGiven", req.FatherGiven, req.FatherGivenFuzziness, true)
	shouldQueries = append(shouldQueries, shouldSubqueries...)
	mustQueries = append(mustQueries, mustSubqueries...)
	shouldSubqueries, mustSubqueries = constructNameQueries("fatherSurname", req.FatherSurname, req.FatherSurnameFuzziness, false)
	shouldQueries = append(shouldQueries, shouldSubqueries...)
	mustQueries = append(mustQueries, mustSubqueries...)
	shouldSubqueries, mustSubqueries = constructNameQueries("motherGiven", req.MotherGiven, req.MotherGivenFuzziness, true)
	shouldQueries = append(shouldQueries, shouldSubqueries...)
	mustQueries = append(mustQueries, mustSubqueries...)
	shouldSubqueries, mustSubqueries = constructNameQueries("motherSurname", req.MotherSurname, req.MotherSurnameFuzziness, false)
	shouldQueries = append(shouldQueries, shouldSubqueries...)
	mustQueries = append(mustQueries, mustSubqueries...)
	shouldSubqueries, mustSubqueries = constructNameQueries("spouseGiven", req.SpouseGiven, req.SpouseGivenFuzziness, true)
	shouldQueries = append(shouldQueries, shouldSubqueries...)
	mustQueries = append(mustQueries, mustSubqueries...)
	shouldSubqueries, mustSubqueries = constructNameQueries("spouseSurname", req.SpouseSurname, req.SpouseSurnameFuzziness, false)
	shouldQueries = append(shouldQueries, shouldSubqueries...)
	mustQueries = append(mustQueries, mustSubqueries...)
	shouldSubqueries, mustSubqueries = constructNameQueries("otherGiven", req.OtherGiven, req.OtherGivenFuzziness, true)
	shouldQueries = append(shouldQueries, shouldSubqueries...)
	mustQueries = append(mustQueries, mustSubqueries...)
	shouldSubqueries, mustSubqueries = constructNameQueries("otherSurname", req.OtherSurname, req.OtherSurnameFuzziness, false)
	shouldQueries = append(shouldQueries, shouldSubqueries...)
	mustQueries = append(mustQueries, mustSubqueries...)

	// events
	shouldSubqueries, mustSubqueries = constructDateQueries("birthYear", "birthDateStd", req.BirthDate, req.BirthDateFuzziness)
	shouldQueries = append(shouldQueries, shouldSubqueries...)
	mustQueries = append(mustQueries, mustSubqueries...)
	shouldSubqueries, mustSubqueries = constructPlaceQueries("birthPlace", req.BirthPlace, req.BirthPlaceFuzziness)
	shouldQueries = append(shouldQueries, shouldSubqueries...)
	mustQueries = append(mustQueries, mustSubqueries...)
	shouldSubqueries, mustSubqueries = constructDateQueries("marriageYear", "marriageDateStd", req.MarriageDate, req.MarriageDateFuzziness)
	shouldQueries = append(shouldQueries, shouldSubqueries...)
	mustQueries = append(mustQueries, mustSubqueries...)
	shouldSubqueries, mustSubqueries = constructPlaceQueries("marriagePlace", req.MarriagePlace, req.MarriagePlaceFuzziness)
	shouldQueries = append(shouldQueries, shouldSubqueries...)
	mustQueries = append(mustQueries, mustSubqueries...)
	shouldSubqueries, mustSubqueries = constructDateQueries("residenceYear", "residenceDateStd", req.ResidenceDate, req.ResidenceDateFuzziness)
	shouldQueries = append(shouldQueries, shouldSubqueries...)
	mustQueries = append(mustQueries, mustSubqueries...)
	shouldSubqueries, mustSubqueries = constructPlaceQueries("residencePlace", req.ResidencePlace, req.ResidencePlaceFuzziness)
	shouldQueries = append(shouldQueries, shouldSubqueries...)
	mustQueries = append(mustQueries, mustSubqueries...)
	shouldSubqueries, mustSubqueries = constructDateQueries("deathYear", "deathDateStd", req.DeathDate, req.DeathDateFuzziness)
	shouldQueries = append(shouldQueries, shouldSubqueries...)
	mustQueries = append(mustQueries, mustSubqueries...)
	shouldSubqueries, mustSubqueries = constructPlaceQueries("deathPlace", req.DeathPlace, req.DeathPlaceFuzziness)
	shouldQueries = append(shouldQueries, shouldSubqueries...)
	mustQueries = append(mustQueries, mustSubqueries...)

	// any date
	if len(req.AnyDate) > 0 {
		var anyShouldQueries []Query
		var anyMustQueries []Query
		shouldSubqueries, mustSubqueries = constructDateQueries("birthYear", "birthDateStd", req.AnyDate, req.AnyDateFuzziness)
		anyShouldQueries = append(anyShouldQueries, shouldSubqueries...)
		anyMustQueries = append(anyMustQueries, mustSubqueries...)
		shouldSubqueries, mustSubqueries = constructDateQueries("marriageYear", "marriageDateStd", req.AnyDate, req.AnyDateFuzziness)
		anyShouldQueries = append(anyShouldQueries, shouldSubqueries...)
		anyMustQueries = append(anyMustQueries, mustSubqueries...)
		shouldSubqueries, mustSubqueries = constructDateQueries("residenceYear", "residenceDateStd", req.AnyDate, req.AnyDateFuzziness)
		anyShouldQueries = append(anyShouldQueries, shouldSubqueries...)
		anyMustQueries = append(anyMustQueries, mustSubqueries...)
		shouldSubqueries, mustSubqueries = constructDateQueries("deathYear", "deathDateStd", req.AnyDate, req.AnyDateFuzziness)
		anyShouldQueries = append(anyShouldQueries, shouldSubqueries...)
		anyMustQueries = append(anyMustQueries, mustSubqueries...)
		shouldSubqueries, mustSubqueries = constructDateQueries("otherYear", "otherDateStd", req.AnyDate, req.AnyDateFuzziness)
		anyShouldQueries = append(anyShouldQueries, shouldSubqueries...)
		anyMustQueries = append(anyMustQueries, mustSubqueries...)
		if len(anyShouldQueries) > 0 {
			shouldQueries = append(shouldQueries, Query{
				DisMax: &DisMaxQuery{
					Queries: anyShouldQueries,
				},
			})
		}
		if len(anyMustQueries) > 0 {
			mustQueries = append(mustQueries, Query{
				DisMax: &DisMaxQuery{
					Queries: anyMustQueries,
				},
			})
		}
	}

	// any place
	if len(req.AnyPlace) > 0 {
		var anyShouldQueries []Query
		var anyMustQueries []Query
		shouldSubqueries, mustSubqueries = constructPlaceQueries("birthPlace", req.AnyPlace, req.AnyPlaceFuzziness)
		anyShouldQueries = append(anyShouldQueries, shouldSubqueries...)
		anyMustQueries = append(anyMustQueries, mustSubqueries...)
		shouldSubqueries, mustSubqueries = constructPlaceQueries("marriagePlace", req.AnyPlace, req.AnyPlaceFuzziness)
		anyShouldQueries = append(anyShouldQueries, shouldSubqueries...)
		anyMustQueries = append(anyMustQueries, mustSubqueries...)
		shouldSubqueries, mustSubqueries = constructPlaceQueries("residencePlace", req.AnyPlace, req.AnyPlaceFuzziness)
		anyShouldQueries = append(anyShouldQueries, shouldSubqueries...)
		anyMustQueries = append(anyMustQueries, mustSubqueries...)
		shouldSubqueries, mustSubqueries = constructPlaceQueries("deathPlace", req.AnyPlace, req.AnyPlaceFuzziness)
		anyShouldQueries = append(anyShouldQueries, shouldSubqueries...)
		anyMustQueries = append(anyMustQueries, mustSubqueries...)
		shouldSubqueries, mustSubqueries = constructPlaceQueries("otherPlace", req.AnyPlace, req.AnyPlaceFuzziness)
		anyShouldQueries = append(anyShouldQueries, shouldSubqueries...)
		anyMustQueries = append(anyMustQueries, mustSubqueries...)
		if len(anyShouldQueries) > 0 {
			shouldQueries = append(shouldQueries, Query{
				DisMax: &DisMaxQuery{
					Queries: anyShouldQueries,
				},
			})
		}
		if len(anyMustQueries) > 0 {
			mustQueries = append(mustQueries, Query{
				DisMax: &DisMaxQuery{
					Queries: anyMustQueries,
				},
			})
		}
	}

	// other
	mustQueries = append(mustQueries, constructTextQueries("keywords", req.Keywords)...)

	// filters
	filterQueries = append(filterQueries, constructFilterQueries("category", req.Category)...)
	filterQueries = append(filterQueries, constructFilterQueries("collection", req.Collection)...)
	filterQueries = append(filterQueries, constructFilterQueries("collectionPlace1", req.CollectionPlace1)...)
	filterQueries = append(filterQueries, constructFilterQueries("collectionPlace2", req.CollectionPlace2)...)
	filterQueries = append(filterQueries, constructFilterQueries("collectionPlace3", req.CollectionPlace3)...)

	// facets
	addTermsAgg(aggs, "category", req.CategoryFacet)
	addTermsAgg(aggs, "collection", len(req.Category) > 0 && req.CollectionFacet)
	addTermsAgg(aggs, "collectionPlace1", req.CollectionPlace1Facet)
	addTermsAgg(aggs, "collectionPlace2", len(req.CollectionPlace1) > 0 && req.CollectionPlace2Facet)
	addTermsAgg(aggs, "collectionPlace3", len(req.CollectionPlace1) > 0 && len(req.CollectionPlace2) > 0 && req.CollectionPlace3Facet)
	if len(aggs) == 0 {
		aggs = nil
	}

	from := req.From
	if from > MaxFrom {
		from = MaxFrom
	}
	size := req.Size
	if size > MaxSize {
		size = MaxSize
	} else if size <= 0 {
		size = DefaultSize
	}

	return &Search{
		Query: Query{
			Bool: &BoolQuery{
				Must:   mustQueries,
				Should: shouldQueries,
				Filter: filterQueries,
			},
		},
		Aggs: aggs,
		From: from,
		Size: size,
	}
}

// TODO learn the best boost values
const exactNameBoost = 1.0
const narrowNameBoost = 0.8
const wildcardNameBoost = 0.7
const broadNameBoost = 0.6
const initialNameBoost = 0.4
const fuzzyNameBoost = 0.2

func constructNameQueries(label, value string, fuzziness int, isGiven bool) ([]Query, []Query) {
	if len(value) == 0 {
		return nil, nil
	}
	var queries []Query

	for _, v := range splitWord(value) {
		if strings.ContainsAny(v, "*?") {
			v := stdtext.AsciiFold(strings.ToLower(v))

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

		if fuzziness == FuzzyNameExact {
			queries = append(queries, exactQuery)
			continue
		}

		subqueries := []Query{exactQuery}

		if fuzziness == FuzzyNameDefault || fuzziness&FuzzyNameAlternate > 0 {
			// TODO alternate spellings
		}
		// TODO choose the best coders for broad and narrow
		if fuzziness == FuzzyNameDefault || fuzziness&FuzzyNameSoundsLikeNarrow > 0 {
			subqueries = append(subqueries, Query{
				Match: map[string]MatchQuery{
					label + ".narrow": {
						Query: v,
						Boost: narrowNameBoost,
					},
				},
			})
		}
		if fuzziness == FuzzyNameDefault || fuzziness&FuzzyNameSoundsLikeBroad > 0 {
			subqueries = append(subqueries, Query{
				Match: map[string]MatchQuery{
					label + ".broad": {
						Query: v,
						Boost: broadNameBoost,
					},
				},
			})
		}
		if fuzziness == FuzzyNameDefault || fuzziness&FuzzyNameLevenshtein > 0 {
			std := stdtext.AsciiFold(strings.ToLower(v))
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
		if fuzziness == FuzzyNameDefault || fuzziness&FuzzyNameInitials > 0 && isGiven {
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

	if fuzziness == FuzzyNameDefault {
		return queries, nil
	} else {
		return nil, queries
	}
}

const exactYearBoost = 0.7
const rangeYearBoost = 0.3

func constructDateQueries(yearLabel, dateLabel, value string, fuzziness int) ([]Query, []Query) {
	if len(value) != 4 {
		return nil, nil
	}

	year, err := strconv.Atoi(value)
	if err != nil {
		return nil, nil
	}

	query := Query{
		Term: map[string]TermQuery{
			yearLabel: {
				Value: value,
				Boost: exactYearBoost,
			},
		},
	}

	if fuzziness == FuzzyDateDefault || fuzziness > FuzzyDateExact {
		var yrRange int
		switch fuzziness {
		case FuzzyDateDefault:
			yrRange = 5
		case FuzzyDateOne:
			yrRange = 1
		case FuzzyDateTwo:
			yrRange = 2
		case FuzzyDateFive:
			yrRange = 5
		case FuzzyDateTen:
			yrRange = 10
		}
		query = Query{
			DisMax: &DisMaxQuery{
				Queries: []Query{query, {
					Range: map[string]RangeQuery{
						yearLabel: {
							GTE:   year - yrRange,
							LTE:   year + yrRange,
							Boost: rangeYearBoost,
						},
					},
				}},
			},
		}
	}

	if fuzziness == FuzzyDateDefault {
		return []Query{query}, nil
	} else {
		return nil, []Query{query}
	}
}

const exactPlaceBoost = 1.0
const wildcardPlaceBoost = 0.7
const fuzzyPlaceBoost = 0.2
const levelPlaceBoost = 0.2

func constructPlaceQueries(label, value string, fuzziness int) ([]Query, []Query) {
	if len(value) == 0 {
		return nil, nil
	}

	// support wildcards within words or ~word, which means to fuzzy-match word
	if strings.ContainsAny(value, "~*?") {
		var queries []Query
		for _, v := range splitWord(value) {
			v := stdtext.AsciiFold(strings.ToLower(v))
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

		if fuzziness == FuzzyPlaceDefault {
			return queries, nil
		} else {
			return nil, queries
		}
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
					Value: strings.Join(levels, ","),
					Boost: exactPlaceBoost,
				},
			},
		},
		{
			Term: map[string]TermQuery{
				fmt.Sprintf("%s%d", label, len(levels)): {
					Value: strings.Join(levels, ",") + ",",
					Boost: exactPlaceBoost,
				},
			},
		},
	}

	if fuzziness == FuzzyPlaceDefault || fuzziness&FuzzyPlaceHigherJurisdictions > 0 {
		for i := 1; i < len(levels); i++ {
			// don't match on just "United States"
			if i == 1 && levels[0] == "United States" {
				continue
			}
			queries = append(queries, Query{
				Term: map[string]TermQuery{
					fmt.Sprintf("%s%d", label, i): {
						Value: strings.Join(levels[0:i], ","),
						Boost: float32(i) * levelPlaceBoost,
					},
				},
			})
		}
	}

	// TODO include nearby places (lat and lon)
	if fuzziness == FuzzyPlaceDefault || fuzziness&FuzzyPlaceNearby > 0 {

	}

	if len(queries) > 1 {
		queries = []Query{
			{
				DisMax: &DisMaxQuery{
					Queries: queries,
				},
			},
		}
	}

	if fuzziness == FuzzyPlaceDefault {
		return queries, nil
	} else {
		return nil, queries
	}
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

	// populate events
	events := []model.SearchEvent{}
	for _, eventType := range EventTypes {
		if data[eventType+"Date"] != "" || data[eventType+"Place"] != "" {
			events = append(events, model.SearchEvent{
				Type:  eventType,
				Date:  data[eventType+"Date"],
				Place: data[eventType+"Place"],
			})
		}
	}

	// populate relationships
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

func getDataForRole(mappings []model.CollectionMapping, record *model.Record, role string) map[string]string {
	data := map[string]string{}

	for _, mapping := range mappings {
		// get marriage data for spouse too
		if record.Data[mapping.Header] != "" &&
			(mapping.IxRole == role || (isSpouseRole(mapping.IxRole, role) && isMarriageField(mapping.IxField))) {
			data[mapping.IxField] = record.Data[mapping.Header]
			if strings.HasSuffix(mapping.IxField, "Date") {
				data[mapping.IxField+stddate.StdSuffix] = record.Data[mapping.Header+stddate.StdSuffix]
			} else if strings.HasSuffix(mapping.IxField, "Place") {
				data[mapping.IxField+stdplace.StdSuffix] = record.Data[mapping.Header+stdplace.StdSuffix]
			}
		}
	}
	return data
}

func isMarriageField(field string) bool {
	return field == "marriageDate" || field == "marriagePlace"
}

func isSpouseRole(role1, role2 string) bool {
	switch role1 {
	case "principal":
		return role2 == "spouse"
	case "spouse":
		return role2 == "principal"
	case "father":
		return role2 == "mother"
	case "mother":
		return role2 == "father"
	case "bride":
		return role2 == "groom"
	case "groom":
		return role2 == "bride"
	case "brideFather":
		return role2 == "brideMother"
	case "brideMother":
		return role2 == "brideFather"
	case "groomFather":
		return role2 == "groomMother"
	case "groomMother":
		return role2 == "groomFather"
	}
	return false
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

func getYears(dateParts, dateRange []string) []int {
	years := []int{}
	switch {
	case len(dateParts) == 1:
		year, err := strconv.Atoi(dateParts[0][:4])
		if err != nil {
			break
		}
		years = append(years, year)
	case len(dateParts) == 2 && len(dateRange) == 1:
		firstYear, err := strconv.Atoi(dateParts[0][:4])
		if err != nil {
			break
		}
		years = append(years, firstYear)
		secondYear, err := strconv.Atoi(dateParts[1][:4])
		if err != nil {
			break
		}
		if secondYear != firstYear {
			years = append(years, secondYear)
		}
	case len(dateParts) == 2 && len(dateRange) == 2:
		startYear, err := strconv.Atoi(dateRange[0][:4])
		if err != nil {
			break
		}
		endYear, err := strconv.Atoi(dateRange[1][:4])
		if err != nil {
			break
		}
		for y := startYear; y <= endYear; y++ {
			years = append(years, y)
		}
	}
	return years

}

func getDatesYears(encodedDate string) ([]int, []int, bool) {
	if encodedDate == "" {
		return nil, nil, false
	}
	// parse encoded date
	dateParts := strings.Split(encodedDate, ",")
	var dateRange []string
	if len(dateParts) == 2 {
		dateRange = strings.Split(dateParts[1], "-")
	}

	// get dates
	dates := []int{}
	for i := 0; i < len(dateParts); i++ {
		ymd, err := strconv.Atoi(dateParts[i])
		if err != nil {
			return nil, nil, false
		}
		dates = append(dates, ymd)
		// get just one date for range
		if len(dateRange) == 2 {
			break
		}
	}

	// get years
	years := getYears(dateParts, dateRange)

	return dates, years, true
}

func getPlaceLevels(stdPlace string) []string {
	var stdLevels []string
	if stdPlace == "" {
		return stdLevels
	}
	levels := splitPlace(stdPlace)
	var std string
	for i := len(levels) - 1; i >= 0; i-- {
		std += strings.TrimSpace(levels[i])
		if i > 0 {
			std += ","
		}
		stdLevels = append(stdLevels, std)
	}
	return stdLevels
}

func getPlaceFacets(stdPlace string) []string {
	var stdLevels []string
	if stdPlace == "" {
		return stdLevels
	}
	levels := splitPlace(stdPlace)
	for i := len(levels) - 1; i >= 0; i-- {
		stdLevels = append(stdLevels, strings.TrimSpace(levels[i]))
	}
	return stdLevels
}

var wordRegexp = regexp.MustCompile("[^\\pL*?~]+") // keep ~*? for fuzzy and wildcards
func splitWord(name string) []string {
	return wordRegexp.Split(name, -1)
}

var placeRegexp = regexp.MustCompile("\\s*,\\s*")

func splitPlace(place string) []string {
	return placeRegexp.Split(place, -1)
}

func reverse(s []string) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
