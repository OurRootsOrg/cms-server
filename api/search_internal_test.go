package api

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
					]}}}`,
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
					]}}}`,
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
					]}}}`,
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
					]}}}`,
		},
	}

	for _, test := range tests {
		var search Search
		json.Unmarshal([]byte(test.query), &search)
		result := constructSearchQuery(&test.req)
		bs, _ := json.Marshal(result)
		fmt.Println(string(bs))
		assert.EqualValues(t, search, *result)
	}
}
