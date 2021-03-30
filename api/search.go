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
	"strconv"
	"strings"

	"github.com/ourrootsorg/cms-server/utils"

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

const Public = "PUBLIC"

const MaxFrom = 1000
const MaxSize = 100
const DefaultSize = 10

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

	// can user access this record?
	if collection.PrivacyLevel&model.PrivacyPrivateDetail > 0 {
		userID, err := utils.GetSearchUserIDFromContext(ctx)
		if err != nil || userID == 0 {
			society, err := api.GetSociety(ctx, hitData.SocietyID)
			if err != nil {
				return nil, err
			}
			return &model.SearchHit{
				ID:                 hitData.ID,
				SocietyID:          hitData.SocietyID,
				CollectionName:     collection.Name,
				CollectionType:     collection.CollectionType,
				CollectionID:       collection.ID,
				CollectionLocation: collection.Location,
				Private:            true,
				LoginURL:           society.LoginURL,
			}, nil
		}
	}

	// read household
	householdRecords := []model.SearchRecord{}
	if collection.HouseholdNumberHeader != "" {
		recordHousehold, errs := api.GetRecordHousehold(ctx, recordDetail.Post, recordDetail.Data[collection.HouseholdNumberHeader])
		if errs != nil || recordHousehold == nil || len(recordHousehold.Records) == 0 {
			log.Printf("[ERROR] recordHousehold not found for record %d err=%v\n", recordDetail.ID, errs)
			return nil, errs
		}
		recs, errs := api.GetRecordsByID(ctx, recordHousehold.Records, true)
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
	var searchPerson model.SearchPerson
	if collection.CollectionType == model.CollectionTypeRecords {
		searchPerson = constructRecordSearchPerson(collection.Mappings, hitData.Role, &recordDetail.Record, false)
	} else {
		searchPerson = constructCatalogSearchPerson(collection.Mappings, hitData.Role, &recordDetail.Record, false)
	}
	return &model.SearchHit{
		ID:                 hitData.ID,
		SocietyID:          hitData.SocietyID,
		Person:             searchPerson,
		Record:             constructSearchRecord(collection.Mappings, &recordDetail.Record),
		CollectionName:     collection.Name,
		CollectionType:     collection.CollectionType,
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

// SearchImage returns a signed S3 URL to return an image file
func (api *API) SearchImage(ctx context.Context, imgSocietyID, postID uint32, filePath string, thumbnail bool, expireSeconds int) (*ImageMetadata, error) {
	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := utils.GetSearchUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	// verify specified SocietyID is the same as context society
	if societyID != imgSocietyID {
		err = fmt.Errorf("user society %d does not match image society %d", societyID, imgSocietyID)
		return nil, NewHTTPError(err, http.StatusForbidden)
	}
	// if the user is not logged in, verify the image is not private
	if userID == 0 {
		// read the post
		post, err := api.GetPost(ctx, postID)
		if err != nil {
			return nil, err
		}
		collection, err := api.GetCollection(ctx, post.Collection)
		if err != nil {
			return nil, err
		}
		society, err := api.GetSociety(ctx, imgSocietyID)
		if err != nil {
			return nil, err
		}
		if (thumbnail && collection.PrivacyLevel&model.PrivacyPrivateDetail > 0) ||
			(!thumbnail && collection.PrivacyLevel&model.PrivacyPrivateImages > 0) {
			return &ImageMetadata{
				Private:  true,
				LoginURL: society.LoginURL,
			}, nil
		}
	}

	return api.GetPostImage(ctx, postID, filePath, thumbnail, expireSeconds)
}

// Search
func (api API) Search(ctx context.Context, req *SearchRequest) (*model.SearchResult, error) {
	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := utils.GetSearchUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

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
		if hitData.SocietyID != societyID { // shouldn't be necessary, but just to be sure
			err = fmt.Errorf("search result SocietyID %d doesn't match context society ID %d", hitData.SocietyID, societyID)
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
	records, errs := api.GetRecordsByID(ctx, recordIDs, false)
	if errs != nil {
		return nil, errs
	}
	collections, errs := api.GetCollectionsByID(ctx, collectionIDs, false)
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

		// mask details on private records when the user is not logged in
		maskDetails := collection.PrivacyLevel&model.PrivacyPrivateDetail > 0 && userID == 0
		// construct search hit
		var searchPerson model.SearchPerson
		if collection.CollectionType == model.CollectionTypeRecords {
			searchPerson = constructRecordSearchPerson(collection.Mappings, hitData.Role, &record, maskDetails)
		} else {
			searchPerson = constructCatalogSearchPerson(collection.Mappings, hitData.Role, &record, maskDetails)
		}
		hits = append(hits, model.SearchHit{
			ID:             hitData.ID,
			SocietyID:      hitData.SocietyID,
			Person:         searchPerson,
			CollectionName: collection.Name,
			CollectionType: collection.CollectionType,
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
	if r.Source.SocietyID == 0 {
		msg := fmt.Sprintf("Missing socidyID for ID %s\n", r.ID)
		log.Printf("[ERROR] %s\n", msg)
		return nil, errors.New(msg)
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

func constructRecordSearchPerson(mappings []model.CollectionMapping, role model.Role, record *model.Record, maskDetails bool) model.SearchPerson {
	data := getDataForRole(mappings, record, role)

	// populate events
	events := []model.SearchEvent{}
	if !maskDetails {
		for _, eventType := range model.EventTypes {
			if data[string(eventType)+"Date"] != "" || data[string(eventType)+"Place"] != "" {
				events = append(events, model.SearchEvent{
					Type:  eventType,
					Date:  data[string(eventType)+"Date"],
					Place: data[string(eventType)+"Place"],
				})
			}
		}
	}

	// populate relationships
	relationships := []model.SearchRelationship{}
	if !maskDetails {
		for _, relative := range model.Relatives {
			names := getNames(mappings, record, RelativeRoles[role][relative])
			if len(names) > 0 {
				relationships = append(relationships, model.SearchRelationship{
					Type: relative,
					Name: strings.Join(getNameParts(names, func(name GivenSurname) string { return fmt.Sprintf("%s %s", name.given, name.surname) }), ", "),
				})
			}
		}
	}

	return model.SearchPerson{
		Name:          fmt.Sprintf("%s %s", data["given"], data["surname"]),
		Role:          role,
		Events:        events,
		Relationships: relationships,
	}
}

func constructCatalogSearchPerson(mappings []model.CollectionMapping, role model.Role, record *model.Record, maskDetails bool) model.SearchPerson {
	data := getDataForRole(mappings, record, role)

	// populate events
	events := []model.SearchEvent{}
	if !maskDetails && data[string(model.OtherEvent)+"Place"] != "" {
		events = append(events, model.SearchEvent{
			Place: data[string(model.OtherEvent)+"Place"],
		})
	}

	// populate relationships
	relationships := []model.SearchRelationship{}
	if !maskDetails && data["surname"] != "" {
		relationships = append(relationships, model.SearchRelationship{
			Name: data["surname"],
		})
	}

	author := "Book"
	if data["author"] != "" {
		author = data["author"]
	}
	return model.SearchPerson{
		Name:          data["title"],
		Role:          model.Role(author),
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
