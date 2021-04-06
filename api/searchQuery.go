package api

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/stdtext"
	"github.com/ourrootsorg/cms-server/utils"
)

// SearchRequest contains the possible search request parameters
type SearchRequest struct {
	SocietyID uint32 `schema:"societyId"`
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
	Title    string `schema:"title"`
	Author   string `schema:"author"`
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
	Size   int            `json:"size"`
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
	Query    string  `json:"query"`
	Operator string  `json:"operator,omitempty"`
	Boost    float32 `json:"boost,omitempty"`
}
type RangeQuery struct {
	GTE   int     `json:"gte,omitempty"`
	LTE   int     `json:"lte,omitempty"`
	Boost float32 `json:"boost,omitempty"`
}
type TermQuery struct {
	Value interface{} `json:"value"`
	Boost float32     `json:"boost,omitempty"`
}
type IntegerTermQuery struct {
	Value uint32  `json:"value"`
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

func (api API) constructSearchQuery(ctx context.Context, req *SearchRequest) (*Search, error) {
	var mustQueries []Query
	var shouldQueries []Query
	var filterQueries []Query
	aggs := map[string]Agg{}

	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := utils.GetSearchUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if societyID != req.SocietyID {
		return nil, fmt.Errorf("jwt societyId %d does not match query societyId %d", societyID, req.SocietyID)
	}

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
	mustQueries = append(mustQueries, constructTextQueries("book_title", req.Title)...)
	mustQueries = append(mustQueries, constructTextQueries("book_author", req.Author)...)

	// filters
	filterQueries = append(filterQueries, constructFilterQueries("societyId", float64(societyID))...) // convert to float64 so tests pass
	if userID == 0 {                                                                                  // not signed in
		filterQueries = append(filterQueries, constructFilterQueries("privacy", Public)...)
	}
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
	} else if size < 0 {
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
		if (fuzziness == FuzzyNameDefault || fuzziness&FuzzyNameInitials > 0) && nameType == model.GivenType {
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
	if !strings.ContainsAny(value, "~*?") {
		return []Query{
			{
				Match: map[string]MatchQuery{
					label: {
						Query:    value,
						Operator: "AND",
					},
				},
			},
		}
	}
	// support wildcards within words or ~word, which means to fuzzy-match word
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
					},
				},
			})
			continue
		}
		v = strings.ReplaceAll(v, "~", "")
		if strings.ContainsAny(v, "*?") {
			queries = append(queries, Query{
				Wildcard: map[string]TermQuery{
					label: {
						Value: v,
					},
				},
			})
			continue
		}
		queries = append(queries, Query{
			Term: map[string]TermQuery{
				label: {
					Value: v,
				},
			},
		})
	}
	return queries
}

func constructFilterQueries(label string, value interface{}) []Query {
	strValue, ok := value.(string)
	if ok && len(strValue) == 0 {
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

var wordRegexp = regexp.MustCompile("[^\\pL*?~]+") // keep ~*? for fuzzy and wildcards
func splitWord(name string) []string {
	return wordRegexp.Split(name, -1)
}

func reverse(s []string) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
