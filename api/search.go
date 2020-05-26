package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ourrootsorg/cms-server/model"
)

type SearchRequest struct {
	Given               string `schema:"given"`
	GivenFuzziness      int    `schema:"givenFuzziness"`
	Surname             string `schema:"surname"`
	SurnameFuzziness    int    `schema:"surnameFuzziness"`
	BirthDate           string `schema:"birthDate"`
	BirthDateFuzziness  int    `schema:"birthDateFuzziness"`
	BirthPlace          string `schema:"birthPlace"`
	BirthPlaceFuzziness int    `schema:"birthPlaceFuzziness"`
	Keywords            string `schema:"keywords"`
	// faceting & filtering
	CategoryFacet     bool   `schema:"categoryFacet"`
	Category          string `schema:"category"`
	CollectionFacet   bool   `schema:"collectionFacet"`
	Collection        string `schema:"collection"`
	BirthPlace1Facet  bool   `schema:"birthPlace1Facet"`
	BirthPlace1       string `schema:"birthPlace1"`
	BirthPlace2Facet  bool   `schema:"birthPlace2Facet"`
	BirthPlace2       string `schema:"birthPlace2"`
	BirthPlace3Facet  bool   `schema:"birthPlace3Facet"`
	BirthPlace3       string `schema:"birthPlace3"`
	BirthPlace4Facet  bool   `schema:"birthPlace4Facet"`
	BirthPlace4       string `schema:"birthPlace4"`
	BirthCenturyFacet bool   `schema:"birthCenturyFacet"`
	BirthCentury      string `schema:"birthCentury"`
	BirthDecadeFacet  bool   `schema:"birthDecadeFacet"`
	BirthDecade       string `schema:"birthDecade"`
}

type SearchResult map[string]interface{}

type Search struct {
	Query  Query          `json:"query,omitempty"`
	Aggs   map[string]Agg `json:"aggs,omitempty"`
	Source []string       `json:"_source,omitempty"`
}
type Query struct {
	Bool     *BoolQuery            `json:"bool,omitempty"`
	DisMax   *DisMaxQuery          `json:"dis_max,omitempty"`
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
type MatchQuery struct {
	Query string  `json:"query"`
	Boost float32 `json:"boost,omitempty"`
}
type RangeQuery struct {
	GTE   string  `json:"gte,omitempty"`
	LTE   string  `json:"lte,omitempty"`
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

// Search
func (api API) Search(ctx context.Context, req SearchRequest) (SearchResult, *model.Errors) {
	search := Search{
		Query: Query{
			Bool: &BoolQuery{
				Should: []Query{
					{
						DisMax: &DisMaxQuery{
							Queries: []Query{
								{
									Match: map[string]MatchQuery{
										"given": {
											Query: req.Given,
											Boost: 1.0,
										},
									},
								},
							},
						},
					},
					{
						DisMax: &DisMaxQuery{
							Queries: []Query{
								{
									Match: map[string]MatchQuery{
										"surname": {
											Query: req.Surname,
											Boost: 1.0,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		Aggs: map[string]Agg{
			"category": {
				Terms: &TermsAgg{
					Field: "category",
					Size:  99,
				},
			},
			"birthPlace1": {
				Terms: &TermsAgg{
					Field: "birthPlace1",
					Size:  99,
				},
			},
			"birthCentury": {
				Range: &RangeAgg{
					Field: "birthDecade",
					Keyed: true,
					Ranges: []RangeAggRange{
						{
							Key: "early",
							To:  1900,
						},
						{
							Key:  "1900s",
							From: 1900,
							To:   2000,
						},
						{
							Key:  "2000s",
							From: 2000,
						},
					},
				},
			},
		},
		Source: []string{"given", "surname", "birthDate", "birthPlace", "category", "collection", "lastMod", "keywords"},
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(search); err != nil {
		log.Printf("[ERROR] encoding query %v\n", err)
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	log.Printf("Query=%s\n", string(buf.Bytes()))
	res, err := api.es.Search(
		api.es.Search.WithContext(ctx),
		api.es.Search.WithIndex("records"),
		api.es.Search.WithBody(&buf),
		api.es.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		log.Printf("[ERROR] Search %v", err)
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Printf("Error parsing the response body: %v", err)
			return nil, model.NewErrors(http.StatusInternalServerError, err)
		} else {
			// Print the response status and error information.
			msg := fmt.Sprintf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
			log.Println(msg)
			return nil, model.NewErrors(http.StatusInternalServerError, errors.New(msg))
		}
	}

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Printf("Error parsing the response body: %s\n", err)
		return nil, model.NewErrors(http.StatusInternalServerError, err)
	}
	// Print the response status, number of results, and request duration.
	log.Printf(
		"[%s] %d hits; took: %dms",
		res.Status(),
		int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		int(r["took"].(float64)),
	)
	// Print the ID and document source for each hit.
	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		log.Printf(" * ID=%s, %s", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
	}

	log.Println(strings.Repeat("=", 37))

	return r, nil
}
