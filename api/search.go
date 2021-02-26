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

	"github.com/ourrootsorg/cms-server/utils"

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
	FuzzyNameVariants         = 1 << iota // 2 - variant spellings
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
	SocietyID    uint32 `json:"societyId"`
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
	SocietyID    uint32
	RecordID     uint32
	Role         model.Role
	CollectionID uint32
}

const numWorkers = 5

type GivenSurname struct {
	given   string
	surname string
}

type nameExtractor func(GivenSurname) string

var IndexRoles = map[model.Role]string{
	model.PrincipalRole:   "",
	model.FatherRole:      "f",
	model.MotherRole:      "m",
	model.SpouseRole:      "s",
	model.BrideRole:       "b",
	model.GroomRole:       "g",
	model.BrideFatherRole: "bf",
	model.BrideMotherRole: "bm",
	model.GroomFatherRole: "gf",
	model.GroomMotherRole: "gm",
	model.OtherRole:       "o",
}

var IndexRolesReversed = reverseRoleMap(IndexRoles)

func reverseRoleMap(m map[model.Role]string) map[string]model.Role {
	result := map[string]model.Role{}
	for k, v := range m {
		result[v] = k
	}
	return result
}

var RelativeRoles = map[model.Role]map[model.Relative][]model.Role{
	model.PrincipalRole: {
		model.FatherRelative: {model.FatherRole},
		model.MotherRelative: {model.MotherRole},
		model.SpouseRelative: {model.SpouseRole},
		model.OtherRelative:  {model.BrideRole, model.GroomRole, model.BrideFatherRole, model.BrideMotherRole, model.GroomFatherRole, model.GroomMotherRole, model.OtherRole},
	},
	model.FatherRole: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {model.MotherRole},
		model.OtherRelative:  {model.PrincipalRole, model.SpouseRole, model.BrideRole, model.GroomRole, model.BrideFatherRole, model.BrideMotherRole, model.GroomFatherRole, model.GroomMotherRole, model.OtherRole},
	},
	model.MotherRole: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {model.FatherRole},
		model.OtherRelative:  {model.PrincipalRole, model.SpouseRole, model.BrideRole, model.GroomRole, model.BrideFatherRole, model.BrideMotherRole, model.GroomFatherRole, model.GroomMotherRole, model.OtherRole},
	},
	model.SpouseRole: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {model.PrincipalRole},
		model.OtherRelative:  {model.FatherRole, model.MotherRole, model.BrideRole, model.GroomRole, model.BrideFatherRole, model.BrideMotherRole, model.GroomFatherRole, model.GroomMotherRole, model.OtherRole},
	},
	model.BrideRole: {
		model.FatherRelative: {model.BrideFatherRole},
		model.MotherRelative: {model.BrideMotherRole},
		model.SpouseRelative: {model.GroomRole},
		model.OtherRelative:  {model.PrincipalRole, model.FatherRole, model.MotherRole, model.SpouseRole, model.GroomFatherRole, model.GroomMotherRole, model.OtherRole},
	},
	model.GroomRole: {
		model.FatherRelative: {model.GroomFatherRole},
		model.MotherRelative: {model.GroomMotherRole},
		model.SpouseRelative: {model.BrideRole},
		model.OtherRelative:  {model.PrincipalRole, model.FatherRole, model.MotherRole, model.SpouseRole, model.BrideFatherRole, model.BrideMotherRole, model.OtherRole},
	},
	model.BrideFatherRole: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {model.BrideMotherRole},
		model.OtherRelative:  {model.PrincipalRole, model.FatherRole, model.MotherRole, model.SpouseRole, model.BrideRole, model.GroomRole, model.GroomFatherRole, model.GroomMotherRole, model.OtherRole},
	},
	model.BrideMotherRole: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {model.BrideFatherRole},
		model.OtherRelative:  {model.PrincipalRole, model.FatherRole, model.MotherRole, model.SpouseRole, model.BrideRole, model.GroomRole, model.GroomFatherRole, model.GroomMotherRole, model.OtherRole},
	},
	model.GroomFatherRole: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {model.GroomMotherRole},
		model.OtherRelative:  {model.PrincipalRole, model.FatherRole, model.MotherRole, model.SpouseRole, model.BrideRole, model.GroomRole, model.BrideFatherRole, model.BrideMotherRole, model.OtherRole},
	},
	model.GroomMotherRole: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {model.GroomFatherRole},
		model.OtherRelative:  {model.PrincipalRole, model.FatherRole, model.MotherRole, model.SpouseRole, model.BrideRole, model.GroomRole, model.BrideFatherRole, model.BrideMotherRole, model.OtherRole},
	},
	model.OtherRole: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {},
		model.OtherRelative:  {model.PrincipalRole, model.FatherRole, model.MotherRole, model.SpouseRole, model.BrideRole, model.GroomRole, model.BrideFatherRole, model.BrideMotherRole, model.GroomFatherRole, model.GroomMotherRole},
	},
}

var RelativeRelationshipsToHead = map[model.HouseholdRelToHead]map[model.Relative][]model.HouseholdRelToHead{
	model.HeadRelToHead: {
		model.FatherRelative: {model.FatherRelToHead},
		model.MotherRelative: {model.MotherRelToHead},
		model.SpouseRelative: {model.SpouseRelToHead, model.HusbandRelToHead, model.WifeRelToHead},
		model.OtherRelative:  {model.HeadRelToHead, model.ChildRelToHead, model.SonRelToHead, model.DaughterRelToHead, model.OtherRelToHead},
	},
	model.FatherRelToHead: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {model.MotherRelToHead},
		model.OtherRelative:  {model.HeadRelToHead, model.FatherRelToHead, model.SpouseRelToHead, model.HusbandRelToHead, model.WifeRelToHead, model.ChildRelToHead, model.SonRelToHead, model.DaughterRelToHead, model.OtherRelToHead},
	},
	model.MotherRelToHead: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {model.FatherRelToHead},
		model.OtherRelative:  {model.HeadRelToHead, model.MotherRelToHead, model.SpouseRelToHead, model.HusbandRelToHead, model.WifeRelToHead, model.ChildRelToHead, model.SonRelToHead, model.DaughterRelToHead, model.OtherRelToHead},
	},
	model.SpouseRelToHead: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {model.HeadRelToHead},
		model.OtherRelative:  {model.FatherRelToHead, model.MotherRelToHead, model.SpouseRelToHead, model.HusbandRelToHead, model.WifeRelToHead, model.ChildRelToHead, model.SonRelToHead, model.DaughterRelToHead, model.OtherRelToHead},
	},
	model.HusbandRelToHead: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {model.HeadRelToHead},
		model.OtherRelative:  {model.FatherRelToHead, model.MotherRelToHead, model.SpouseRelToHead, model.HusbandRelToHead, model.WifeRelToHead, model.ChildRelToHead, model.SonRelToHead, model.DaughterRelToHead, model.OtherRelToHead},
	},
	model.WifeRelToHead: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {model.HeadRelToHead},
		model.OtherRelative:  {model.FatherRelToHead, model.MotherRelToHead, model.SpouseRelToHead, model.HusbandRelToHead, model.WifeRelToHead, model.ChildRelToHead, model.SonRelToHead, model.DaughterRelToHead, model.OtherRelToHead},
	},
	model.ChildRelToHead: {
		model.FatherRelative: {model.HeadRelToHead, model.SpouseRelToHead, model.HusbandRelToHead},
		model.MotherRelative: {model.HeadRelToHead, model.SpouseRelToHead, model.WifeRelToHead},
		model.SpouseRelative: {},
		model.OtherRelative:  {model.FatherRelToHead, model.MotherRelToHead, model.ChildRelToHead, model.SonRelToHead, model.DaughterRelToHead, model.OtherRelToHead},
	},
	model.SonRelToHead: {
		model.FatherRelative: {model.HeadRelToHead, model.SpouseRelToHead, model.HusbandRelToHead},
		model.MotherRelative: {model.HeadRelToHead, model.SpouseRelToHead, model.WifeRelToHead},
		model.SpouseRelative: {},
		model.OtherRelative:  {model.FatherRelToHead, model.MotherRelToHead, model.ChildRelToHead, model.SonRelToHead, model.DaughterRelToHead, model.OtherRelToHead},
	},
	model.DaughterRelToHead: {
		model.FatherRelative: {model.HeadRelToHead, model.SpouseRelToHead, model.HusbandRelToHead},
		model.MotherRelative: {model.HeadRelToHead, model.SpouseRelToHead, model.WifeRelToHead},
		model.SpouseRelative: {},
		model.OtherRelative:  {model.FatherRelToHead, model.MotherRelToHead, model.ChildRelToHead, model.SonRelToHead, model.DaughterRelToHead, model.OtherRelToHead},
	},
	model.OtherRelToHead: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {},
		model.OtherRelative:  {model.HeadRelToHead, model.FatherRelToHead, model.MotherRelToHead, model.SpouseRelToHead, model.HusbandRelToHead, model.WifeRelToHead, model.ChildRelToHead, model.SonRelToHead, model.DaughterRelToHead, model.OtherRelToHead},
	},
}

// IndexPost
func (api API) IndexPost(ctx context.Context, post *model.Post) error {
	var countSuccessful uint64

	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		log.Printf("[ERROR] Missing society %v\n", err)
		return err
	}

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

	// read record households for post
	var recordHouseholds []model.RecordHousehold
	if collection.HouseholdNumberHeader != "" {
		recordHouseholds, errs = api.GetRecordHouseholdsForPost(ctx, post.ID)
		if errs != nil {
			log.Printf("[ERROR] GetRecordHouseholdsForPost %v\n", errs)
			return errs
		}
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

	householdRecordsMap := map[string][]*model.Record{}
	if collection.HouseholdNumberHeader != "" {
		householdRecordsMap = getHouseholdRecordsMap(recordHouseholds, records.Records)
	}

	for _, record := range records.Records {
		var householdRecords []*model.Record
		if collection.HouseholdNumberHeader != "" {
			householdRecords = householdRecordsMap[record.Data[collection.HouseholdNumberHeader]]
		}
		err = indexRecord(&record, householdRecords, societyID, post, collection, categories, lastModified, &countSuccessful, bi)
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

func getHouseholdRecordsMap(recordHouseholds []model.RecordHousehold, records []model.Record) map[string][]*model.Record {
	recordsMap := map[uint32]*model.Record{}
	for ix := range records {
		recordsMap[records[ix].ID] = &records[ix]
	}
	result := map[string][]*model.Record{}
	for _, recordHousehold := range recordHouseholds {
		if recordHousehold.Household == "" {
			continue // should never happen
		}
		var records []*model.Record
		for _, recordID := range recordHousehold.Records {
			records = append(records, recordsMap[recordID])
		}
		result[recordHousehold.Household] = records
	}
	return result
}

func indexRecord(record *model.Record, householdRecords []*model.Record, societyID uint32, post *model.Post, collection *model.Collection,
	categories []model.Category, lastModified string, countSuccessful *uint64, bi esutil.BulkIndexer) error {

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
			if role == "principal" {
				log.Printf("[DEBUG] No given name or surname found for record %#v, mappings %#v, role %s",
					record, collection.Mappings, role)
			}
			continue
		}

		// get relatives' names
		for _, relative := range model.Relatives {
			names := getNames(collection.Mappings, record, RelativeRoles[role][relative])
			// include relatives' names from household
			if role == model.PrincipalRole && collection.HouseholdRelationshipHeader != "" && len(householdRecords) > 0 {
				relToHead := stdRelToHead(record.Data[collection.HouseholdRelationshipHeader])
				householdNames := getHouseholdNames(collection.HouseholdRelationshipHeader, collection.GenderHeader,
					collection.Mappings, relative, RelativeRelationshipsToHead[relToHead][relative], record.ID, householdRecords)
				if len(householdNames) > 0 {
					names = append(names, householdNames...)
				}
			}
			givens := unique(getNameParts(names, func(name GivenSurname) string { return name.given }))
			surnames := unique(getNameParts(names, func(name GivenSurname) string { return name.surname }))
			if len(givens) > 0 {
				ixRecord[string(relative)+"Given"] = strings.Join(givens, " ")
			}
			if len(surnames) > 0 {
				ixRecord[string(relative)+"Surname"] = strings.Join(surnames, " ")
			}
		}

		// get events
		for _, eventType := range model.EventTypes {
			if data[string(eventType)+"Date"] != "" {
				dates, years, valid := getDatesYears(data[string(eventType)+"Date_std"])
				if valid {
					ixRecord[string(eventType)+"DateStd"] = dates
					ixRecord[string(eventType)+"Year"] = years
				}
			}
			if data[string(eventType)+"Place"] != "" {
				placeLevels := getPlaceLevels(data[string(eventType)+"Place_std"])
				if len(placeLevels) > 0 {
					ixRecord[string(eventType)+"Place"] = data[string(eventType)+"Place"]
					ixRecord[string(eventType)+"Place1"] = placeLevels[0]
				}
				if len(placeLevels) > 1 {
					ixRecord[string(eventType)+"Place2"] = placeLevels[1]
				}
				if len(placeLevels) > 2 {
					ixRecord[string(eventType)+"Place3"] = placeLevels[2]
				}
				if len(placeLevels) > 3 {
					ixRecord[string(eventType)+"Place4"] = placeLevels[3]
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
		ixRecord["societyId"] = societyID
		if collection.PrivacyLevel < model.PrivacyPrivateSearch {
			ixRecord["privacy"] = "PUBLIC"
		}
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

	// TODO verify hitData.SocietyID == ctx.societyID

	// read record and collection
	// don't include household because we can get the household here more efficiently (we don't have to re-read post and collection)
	recordDetail, errs := api.GetRecord(ctx, false, hitData.RecordID)
	if errs != nil {
		log.Printf("[ERROR] record not found %d\n", hitData.RecordID)
		return nil, errs
	}
	collection, errs := api.GetCollection(ctx, hitData.CollectionID)
	if errs != nil {
		log.Printf("[ERROR] collection not found %d\n", hitData.CollectionID)
		return nil, errs
	}

	// read household
	householdRecords := []model.SearchRecord{}
	if collection.HouseholdNumberHeader != "" {
		recordHousehold, errs := api.GetRecordHousehold(ctx, recordDetail.Post, recordDetail.Data[collection.HouseholdNumberHeader])
		if errs != nil || recordHousehold == nil || len(recordHousehold.Records) == 0 {
			log.Printf("[ERROR] recordHousehold not found for record %d err=%v\n", recordDetail.ID, errs)
			return nil, errs
		}
		recs, errs := api.GetRecordsByID(ctx, recordHousehold.Records)
		if errs != nil {
			log.Printf("[ERROR] household records not found for record %d err=%v\n", recordDetail.ID, errs)
			return nil, errs
		}
		// add household records in order
		for _, recordID := range recordHousehold.Records {
			for _, rec := range recs {
				if rec.ID == recordID {
					householdRecords = append(householdRecords, constructSearchRecord(collection.Mappings, &rec))
					break
				}
			}
		}
	}
	return &model.SearchHit{
		ID:                 hitData.ID,
		SocietyID:          hitData.SocietyID,
		Person:             constructSearchPerson(collection.Mappings, hitData.Role, &recordDetail.Record),
		Record:             constructSearchRecord(collection.Mappings, &recordDetail.Record),
		CollectionName:     collection.Name,
		CollectionID:       collection.ID,
		CollectionLocation: collection.Location,
		Citation:           recordDetail.GetCitation(collection.CitationTemplate),
		PostID:             recordDetail.Post,
		ImagePath:          recordDetail.Data[collection.ImagePathHeader],
		Household:          householdRecords,
	}, nil
}

func (api API) SearchDeleteByPost(ctx context.Context, id uint32) error {
	// post ID is globally unique
	// ID must have been verified to belong to ctx society; we don't include society in the search
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
	// for testing only
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
	// TODO add societyID as a search facet term
	search, err := api.constructSearchQuery(ctx, req)
	if err != nil {
		return nil, err
	}

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
			ID:             hitData.ID,
			Person:         constructSearchPerson(collection.Mappings, hitData.Role, &record),
			CollectionName: collection.Name,
			CollectionID:   collection.ID,
			PostID:         record.Post,
			ImagePath:      record.Data[collection.ImagePathHeader],
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

func (api API) constructSearchQuery(ctx context.Context, req *SearchRequest) (*Search, error) {
	var mustQueries []Query
	var shouldQueries []Query
	var filterQueries []Query
	aggs := map[string]Agg{}

	// name
	shouldGivenQueries, mustGivenQueries, err := api.constructNameQueries(ctx, "given", req.Given, req.GivenFuzziness, model.GivenType)
	if err != nil {
		return nil, err
	}
	shouldSurnameQueries, mustSurnameQueries, err := api.constructNameQueries(ctx, "surname", req.Surname, req.SurnameFuzziness, model.SurnameType)
	if err != nil {
		return nil, err
	}
	if len(shouldGivenQueries) > 0 || len(shouldSurnameQueries) > 0 || len(mustGivenQueries) > 0 || len(mustSurnameQueries) > 0 {
		mustQueries = append(mustQueries, Query{
			Bool: &BoolQuery{
				Must:   append(mustGivenQueries, mustSurnameQueries...),
				Should: append(shouldGivenQueries, shouldSurnameQueries...),
			},
		})
	}

	// relative names
	shouldSubqueries, mustSubqueries, err := api.constructNameQueries(ctx, "fatherGiven", req.FatherGiven, req.FatherGivenFuzziness, model.GivenType)
	if err != nil {
		return nil, err
	}
	shouldQueries = append(shouldQueries, shouldSubqueries...)
	mustQueries = append(mustQueries, mustSubqueries...)
	shouldSubqueries, mustSubqueries, err = api.constructNameQueries(ctx, "fatherSurname", req.FatherSurname, req.FatherSurnameFuzziness, model.SurnameType)
	if err != nil {
		return nil, err
	}
	shouldQueries = append(shouldQueries, shouldSubqueries...)
	mustQueries = append(mustQueries, mustSubqueries...)
	shouldSubqueries, mustSubqueries, err = api.constructNameQueries(ctx, "motherGiven", req.MotherGiven, req.MotherGivenFuzziness, model.GivenType)
	if err != nil {
		return nil, err
	}
	shouldQueries = append(shouldQueries, shouldSubqueries...)
	mustQueries = append(mustQueries, mustSubqueries...)
	shouldSubqueries, mustSubqueries, err = api.constructNameQueries(ctx, "motherSurname", req.MotherSurname, req.MotherSurnameFuzziness, model.SurnameType)
	if err != nil {
		return nil, err
	}
	shouldQueries = append(shouldQueries, shouldSubqueries...)
	mustQueries = append(mustQueries, mustSubqueries...)
	shouldSubqueries, mustSubqueries, err = api.constructNameQueries(ctx, "spouseGiven", req.SpouseGiven, req.SpouseGivenFuzziness, model.GivenType)
	if err != nil {
		return nil, err
	}
	shouldQueries = append(shouldQueries, shouldSubqueries...)
	mustQueries = append(mustQueries, mustSubqueries...)
	shouldSubqueries, mustSubqueries, err = api.constructNameQueries(ctx, "spouseSurname", req.SpouseSurname, req.SpouseSurnameFuzziness, model.SurnameType)
	if err != nil {
		return nil, err
	}
	shouldQueries = append(shouldQueries, shouldSubqueries...)
	mustQueries = append(mustQueries, mustSubqueries...)
	shouldSubqueries, mustSubqueries, err = api.constructNameQueries(ctx, "otherGiven", req.OtherGiven, req.OtherGivenFuzziness, model.GivenType)
	if err != nil {
		return nil, err
	}
	shouldQueries = append(shouldQueries, shouldSubqueries...)
	mustQueries = append(mustQueries, mustSubqueries...)
	shouldSubqueries, mustSubqueries, err = api.constructNameQueries(ctx, "otherSurname", req.OtherSurname, req.OtherSurnameFuzziness, model.SurnameType)
	if err != nil {
		return nil, err
	}
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
	}, nil
}

// TODO learn the best boost values
const exactNameBoost = 1.0
const variantNameBoost = 0.7
const narrowNameBoost = 0.6
const wildcardNameBoost = 0.5
const broadNameBoost = 0.4
const fuzzyNameBoost = 0.3
const initialNameBoost = 0.2

func (api API) constructNameQueries(ctx context.Context, label, value string, fuzziness int, nameType model.NameType) ([]Query, []Query, error) {
	if len(value) == 0 {
		return nil, nil, nil
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

		if fuzziness == FuzzyNameDefault || fuzziness&FuzzyNameVariants > 0 {
			nameVariants, err := api.GetNameVariants(ctx, nameType, stdtext.AsciiFold(strings.ToLower(v)))
			if err != nil {
				if !model.ErrNotFound.Matches(err) {
					return nil, nil, err
				}
				nameVariants = &model.NameVariants{}
			}
			for _, variant := range nameVariants.Variants {
				subqueries = append(subqueries, Query{
					Match: map[string]MatchQuery{
						label: {
							Query: variant,
							Boost: variantNameBoost,
						},
					},
				})
			}
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
		if fuzziness == FuzzyNameDefault || fuzziness&FuzzyNameInitials > 0 && nameType == model.GivenType {
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
		return queries, nil, nil
	} else {
		return nil, queries, nil
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
	role := model.PrincipalRole
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
		SocietyID:    r.Source.SocietyID,
		RecordID:     uint32(rid),
		Role:         role,
		CollectionID: r.Source.CollectionID,
	}, nil
}

func constructSearchPerson(mappings []model.CollectionMapping, role model.Role, record *model.Record) model.SearchPerson {
	data := getDataForRole(mappings, record, role)

	// populate events
	events := []model.SearchEvent{}
	for _, eventType := range model.EventTypes {
		if data[string(eventType)+"Date"] != "" || data[string(eventType)+"Place"] != "" {
			events = append(events, model.SearchEvent{
				Type:  eventType,
				Date:  data[string(eventType)+"Date"],
				Place: data[string(eventType)+"Place"],
			})
		}
	}

	// populate relationships
	relationships := []model.SearchRelationship{}
	for _, relative := range model.Relatives {
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

func getDataForRole(mappings []model.CollectionMapping, record *model.Record, role model.Role) map[string]string {
	data := map[string]string{}

	for _, mapping := range mappings {
		// get marriage data for spouse too
		if record.Data[mapping.Header] != "" &&
			(mapping.IxRole == string(role) || (isSpouseRole(mapping.IxRole, role) && isMarriageField(mapping.IxField))) {
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

func isSpouseRole(role1 string, role2 model.Role) bool {
	switch role1 {
	case string(model.PrincipalRole):
		return role2 == model.SpouseRole
	case string(model.SpouseRole):
		return role2 == model.PrincipalRole
	case string(model.FatherRole):
		return role2 == model.MotherRole
	case string(model.MotherRole):
		return role2 == model.FatherRole
	case string(model.BrideRole):
		return role2 == model.GroomRole
	case string(model.GroomRole):
		return role2 == model.BrideRole
	case string(model.BrideFatherRole):
		return role2 == model.BrideMotherRole
	case string(model.BrideMotherRole):
		return role2 == model.BrideFatherRole
	case string(model.GroomFatherRole):
		return role2 == model.GroomMotherRole
	case string(model.GroomMotherRole):
		return role2 == model.GroomFatherRole
	}
	return false
}

func getNames(mappings []model.CollectionMapping, record *model.Record, roles []model.Role) []GivenSurname {
	names := []GivenSurname{}

	for _, role := range roles {
		names = append(names, getNamesForRole(role, mappings, record)...)
	}
	return names
}

func getNamesForRole(role model.Role, mappings []model.CollectionMapping, record *model.Record) []GivenSurname {
	var names []GivenSurname

	var givens []string
	var surnames []string
	for _, mapping := range mappings {
		if mapping.IxRole == string(role) {
			if mapping.IxField == "given" && record.Data[mapping.Header] != "" {
				givens = append(givens, record.Data[mapping.Header])
			}
			if mapping.IxField == "surname" && record.Data[mapping.Header] != "" {
				surnames = append(surnames, record.Data[mapping.Header])
			}
		}
	}
	if len(givens) > 0 || len(surnames) > 0 {
		givens = unique(givens)
		surnames = unique(surnames)
		names = append(names, GivenSurname{
			given:   strings.Join(givens, " "),
			surname: strings.Join(surnames, " "),
		})
	}
	return names
}

func getHouseholdNames(relToHeadHeader, genderHeader string, mappings []model.CollectionMapping, relative model.Relative,
	relsToHead []model.HouseholdRelToHead, recordID uint32, householdRecords []*model.Record) []GivenSurname {

	names := []GivenSurname{}

	// for each record in household that is not this record
	for _, record := range householdRecords {
		if record.ID == recordID {
			continue
		}

		// if the record's relationship to head is in relsToHead, get the names
		recordRelToHead := stdRelToHead(record.Data[relToHeadHeader])
		found := false
		for _, relToHead := range relsToHead {
			if recordRelToHead == relToHead {
				found = true
				break
			}
		}
		if !found {
			continue
		}

		// if the relative is father or mother and the relationship is head or spouse, make sure the gender doesn't disagree
		if (relative == model.FatherRelative || relative == model.MotherRelative) &&
			(recordRelToHead == model.HeadRelToHead || recordRelToHead == model.SpouseRelToHead) {
			recordGender := stdGender(record.Data[genderHeader])
			if (relative == model.FatherRelative && recordGender == model.GenderFemale) ||
				(relative == model.MotherRelative && recordGender == model.GenderMale) {
				continue
			}
		}

		// get names
		names = append(names, getNamesForRole(model.PrincipalRole, mappings, record)...)
	}
	return names
}

func stdRelToHead(relToHead string) model.HouseholdRelToHead {
	relToHead = strings.ToLower(relToHead)
	for _, stdRelToHead := range model.HouseholdRelsToHead {
		if relToHead == string(stdRelToHead) {
			return stdRelToHead
		}
	}
	return model.OtherRelToHead
}

func stdGender(gender string) model.Gender {
	gender = strings.ToLower(gender)
	if strings.HasPrefix(gender, "f") {
		return model.GenderFemale
	}
	if strings.HasPrefix(gender, "m") {
		return model.GenderMale
	}
	return model.GenderOther
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

func unique(arr []string) []string {
	var result []string
OUTER:
	for _, s := range arr {
		for _, res := range result {
			if res == s {
				continue OUTER
			}
		}
		result = append(result, s)
	}
	return result
}
