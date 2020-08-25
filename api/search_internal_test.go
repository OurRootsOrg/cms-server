package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDatesYears(t *testing.T) {
	tests := []struct {
		encodedDate string
		dates       []int
		years       []int
	}{
		{
			encodedDate: "19010319",
			dates:       []int{19010319},
			years:       []int{1901},
		},
		{
			encodedDate: "19010319,19010419",
			dates:       []int{19010319, 19010419},
			years:       []int{1901},
		},
		{
			encodedDate: "19010319,19020419",
			dates:       []int{19010319, 19020419},
			years:       []int{1901, 1902},
		},
		{
			encodedDate: "19010319,18990101-19011231",
			dates:       []int{19010319},
			years:       []int{1899, 1900, 1901},
		},
	}

	for i, test := range tests {
		dates, years, valid := getDatesYears(test.encodedDate)
		assert.True(t, valid, i)
		assert.Equal(t, test.dates, dates, i)
		assert.Equal(t, test.years, years, i)
	}
}

func TestGetPlaceLevels(t *testing.T) {
	tests := []struct {
		place  string
		levels []string
	}{
		{
			place:  "United States",
			levels: []string{"United States"},
		},
		{
			place:  "Alabama, United States",
			levels: []string{"United States,", "United States,Alabama"},
		},
		{
			place:  "Autauga, Alabama, United States",
			levels: []string{"United States,", "United States,Alabama,", "United States,Alabama,Autauga"},
		},
	}

	for _, test := range tests {
		levels := getPlaceLevels(test.place)
		assert.Equal(t, test.levels, levels)
	}
}

func TestGetPlaceFacets(t *testing.T) {
	tests := []struct {
		place  string
		levels []string
	}{
		{
			place:  "United States",
			levels: []string{"United States"},
		},
		{
			place:  "Alabama, United States",
			levels: []string{"United States", "Alabama"},
		},
		{
			place:  "Autauga, Alabama, United States",
			levels: []string{"United States", "Alabama", "Autauga"},
		},
	}

	for _, test := range tests {
		levels := getPlaceFacets(test.place)
		assert.Equal(t, test.levels, levels)
	}
}

func TestSearchQuery(t *testing.T) {
	tests := []struct {
		req   SearchRequest
		query string
	}{
		{
			req: SearchRequest{
				Given:            "Fred",
				Surname:          "Flintstone",
				SurnameFuzziness: FuzzyNameExact,
			},
			query: `{"query":{"bool":{"must":[
					  {"bool":{"must":[
						{"match":{"surname":{"query":"Flintstone","boost":1}}}
					  ],"should":[
						{"dis_max":{"queries":[
						  {"match":{"given":{"query":"Fred","boost":1}}},
                      	  {"match":{"given.narrow":{"query":"Fred","boost":0.8}}},
                      	  {"match":{"given.broad":{"query":"Fred","boost":0.6}}},
                      	  {"fuzzy":{"given":{"value":"fred","fuzziness":"AUTO","rewrite":"constant_score_boolean","boost":0.2}}},
                      	  {"match":{"given":{"query":"F","boost":0.4}}}
                    	]}}
					  ]}}
					]}},"from":0,"size":10}`,
		},
		{
			req: SearchRequest{
				Given:            "Fred",
				GivenFuzziness:   FuzzyNameExact | FuzzyNameSoundsLikeNarrow,
				Surname:          "Flintstone",
				SurnameFuzziness: FuzzyNameExact,
			},
			query: `{"query":{"bool":{"must":[
					  {"bool":{"must":[
						{"dis_max":{"queries":[
						  {"match":{"given":{"query":"Fred","boost":1}}},
                      	  {"match":{"given.narrow":{"query":"Fred","boost":0.8}}}
                    	]}},
						{"match":{"surname":{"query":"Flintstone","boost":1}}}
					  ]}}
					]}},"from":0,"size":10}`,
		},
		{
			req: SearchRequest{
				Surname:            "Flintstone",
				SurnameFuzziness:   FuzzyNameExact,
				BirthDate:          "1900",
				BirthDateFuzziness: FuzzyDateTwo,
				DeathDate:          "1995",
			},
			query: `{"query":{"bool":{"must":[
					  {"bool":{"must":[
						{"match":{"surname":{"query":"Flintstone","boost":1}}}
					  ]}},
					  {"dis_max":{"queries":[
						{"term":{"birthYear":{"value":"1900","boost":0.7}}},
						{"range":{"birthYear":{"gte":1898,"lte":1902,"boost":0.3}}}
					  ]}}
					],"should":[
					  {"dis_max":{"queries":[
						{"term":{"deathYear":{"value":"1995","boost":0.7}}},
						{"range":{"deathYear":{"gte":1990,"lte":2000,"boost":0.3}}}
					  ]}}
					]}},"from":0,"size":10}`,
		},
		{
			req: SearchRequest{
				Surname:            "Flintstone",
				SurnameFuzziness:   FuzzyNameExact,
				BirthDate:          "1900",
				BirthDateFuzziness: FuzzyDateOne,
				DeathDate:          "1995",
			},
			query: `{"query":{"bool":{"must":[
					  {"bool":{"must":[
						{"match":{"surname":{"query":"Flintstone","boost":1}}}
					  ]}},
					  {"dis_max":{"queries":[
						{"term":{"birthYear":{"value":"1900","boost":0.7}}},
						{"range":{"birthYear":{"gte":1899,"lte":1901,"boost":0.3}}}
					  ]}}
					],"should":[
					  {"dis_max":{"queries":[
						{"term":{"deathYear":{"value":"1995","boost":0.7}}},
						{"range":{"deathYear":{"gte":1990,"lte":2000,"boost":0.3}}}
					  ]}}
					]}},"from":0,"size":10}`,
		},
		{
			req: SearchRequest{
				Surname:          "Flintstone",
				SurnameFuzziness: FuzzyNameExact,
				BirthPlace:       "Autauga, Alabama, United States",
			},
			query: `{"query":{"bool":{"must":[
					  {"bool":{"must":[
						{"match":{"surname":{"query":"Flintstone","boost":1}}}
					  ]}}
					],"should":[
					  {"dis_max":{"queries":[
						{"term":{"birthPlace3":{"value":"United States,Alabama,Autauga","boost":1.0}}},
						{"term":{"birthPlace3":{"value":"United States,Alabama,Autauga,","boost":1.0}}},
						{"term":{"birthPlace2":{"value":"United States,Alabama","boost":0.4}}}
					  ]}}
					]}},"from":0,"size":10}`,
		},
		{
			req: SearchRequest{
				Surname:             "Flintstone",
				SurnameFuzziness:    FuzzyNameExact,
				BirthPlace:          "Autauga, Alabama, United States",
				BirthPlaceFuzziness: FuzzyPlaceExact,
			},
			query: `{"query":{"bool":{"must":[
					  {"bool":{"must":[
						{"match":{"surname":{"query":"Flintstone","boost":1}}}
					  ]}},
					  {"dis_max":{"queries":[
						{"term":{"birthPlace3":{"value":"United States,Alabama,Autauga","boost":1.0}}},
						{"term":{"birthPlace3":{"value":"United States,Alabama,Autauga,","boost":1.0}}}
					  ]}}
					]}},"from":0,"size":10}`,
		},
		{
			req: SearchRequest{
				Surname:               "Flintstone",
				SurnameFuzziness:      FuzzyNameExact,
				CollectionPlace1:      "United States",
				CollectionPlace2Facet: true,
			},
			query: `{"query":{"bool":{"must":[
					  {"bool":{"must":[
						{"match":{"surname":{"query":"Flintstone","boost":1}}}
					  ]}}
					],
                    "filter":[{"term":{"collectionPlace1":{"value":"United States"}}}]
					}},
					"aggs":{"collectionPlace2":{"terms":{"field":"collectionPlace2","size":250}}},
					"from":0,"size":10}`,
		},
	}

	for i, test := range tests {
		var search Search
		json.Unmarshal([]byte(test.query), &search)
		result := constructSearchQuery(&test.req)
		//bs, _ := json.Marshal(result)
		//fmt.Printf("%d expected=%s\n", i, test.query)
		//fmt.Printf("%d actual  =%s\n", i, string(bs))
		assert.EqualValues(t, search, *result, i)
	}
}
