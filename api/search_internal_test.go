package api

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"testing"

	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/persist"
	"gocloud.dev/postgres"

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

func TestGetHouseholdNames(t *testing.T) {
	relToHeadHeader := "Relationship"
	genderHeader := "Sex"
	mappings := []model.CollectionMapping{
		{
			Header:  "Given name",
			IxRole:  "principal",
			IxField: "given",
		},
		{
			Header:  "Surname",
			IxRole:  "principal",
			IxField: "surname",
		},
	}
	records := []*model.Record{
		{
			ID: 1,
			RecordIn: model.RecordIn{
				RecordBody: model.RecordBody{
					Data: map[string]string{
						"Given name":   "Fred",
						"Surname":      "Flintstone",
						"Relationship": "head",
						"Sex":          "M",
					},
				},
			},
		},
		{
			ID: 2,
			RecordIn: model.RecordIn{
				RecordBody: model.RecordBody{
					Data: map[string]string{
						"Given name":   "Wilma",
						"Surname":      "Flintstone",
						"Relationship": "spouse",
						"Sex":          "F",
					},
				},
			},
		},
		{
			ID: 3,
			RecordIn: model.RecordIn{
				RecordBody: model.RecordBody{
					Data: map[string]string{
						"Given name":   "Pebbles",
						"Surname":      "Flintstone",
						"Relationship": "daughter",
						"Sex":          "F",
					},
				},
			},
		},
		{
			ID: 4,
			RecordIn: model.RecordIn{
				RecordBody: model.RecordBody{
					Data: map[string]string{
						"Given name":   "Pearl",
						"Surname":      "Slaghoople",
						"Relationship": "mother-in-law",
						"Sex":          "F",
					},
				},
			},
		},
	}

	tests := []struct {
		relToHead model.HouseholdRelToHead
		relative  model.Relative
		recordID  uint32
		names     []GivenSurname
	}{
		{
			relToHead: model.HeadRelToHead,
			relative:  model.SpouseRelative,
			recordID:  1,
			names:     []GivenSurname{{given: "Wilma", surname: "Flintstone"}},
		},
		{
			relToHead: model.HeadRelToHead,
			relative:  model.OtherRelative,
			recordID:  1,
			names:     []GivenSurname{{given: "Pebbles", surname: "Flintstone"}, {given: "Pearl", surname: "Slaghoople"}},
		},
		{
			relToHead: model.HeadRelToHead,
			relative:  model.FatherRelative,
			recordID:  1,
			names:     []GivenSurname{},
		},
		{
			relToHead: model.WifeRelToHead,
			relative:  model.SpouseRelative,
			recordID:  2,
			names:     []GivenSurname{{given: "Fred", surname: "Flintstone"}},
		},
		{
			relToHead: model.WifeRelToHead,
			relative:  model.OtherRelative,
			recordID:  2,
			names:     []GivenSurname{{given: "Pebbles", surname: "Flintstone"}, {given: "Pearl", surname: "Slaghoople"}},
		},
		{
			relToHead: model.DaughterRelToHead,
			relative:  model.FatherRelative,
			recordID:  3,
			names:     []GivenSurname{{given: "Fred", surname: "Flintstone"}},
		},
		{
			relToHead: model.DaughterRelToHead,
			relative:  model.MotherRelative,
			recordID:  3,
			names:     []GivenSurname{{given: "Wilma", surname: "Flintstone"}},
		},
		{
			relToHead: model.DaughterRelToHead,
			relative:  model.OtherRelative,
			recordID:  3,
			names:     []GivenSurname{{given: "Pearl", surname: "Slaghoople"}},
		},
	}

	for _, test := range tests {
		names := getHouseholdNames(relToHeadHeader, genderHeader, mappings, test.relative,
			RelativeRelationshipsToHead[test.relToHead][test.relative], test.recordID, records)
		assert.Equal(t, test.names, names)
	}
}

func TestSearchQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL != "" {
		db, err := postgres.Open(context.TODO(), databaseURL)
		if err != nil {
			log.Fatalf("Error opening database connection: %v\n  DATABASE_URL: %s",
				err,
				databaseURL,
			)
		}
		p := persist.NewPostgresPersister(db)
		doInternalSearchTests(t, p)
	}
	// TODO implement
	//dynamoDBTableName := os.Getenv("DYNAMODB_TEST_TABLE_NAME")
	//if dynamoDBTableName != "" {
	//	config := aws.Config{
	//		Region:      aws.String("us-east-1"),
	//		Endpoint:    aws.String("http://localhost:18000"),
	//		DisableSSL:  aws.Bool(true),
	//		Credentials: credentials.NewStaticCredentials("ACCESS_KEY", "SECRET", ""),
	//	}
	//	sess, err := session.NewSession(&config)
	//	assert.NoError(t, err)
	//	p, err := dynamo.NewPersister(sess, dynamoDBTableName)
	//	assert.NoError(t, err)
	//	doInternalSearchTests(t, p)
	//}
}

func doInternalSearchTests(t *testing.T,
	nameP model.NamePersister,
) {
	ctx := context.TODO()
	testApi, err := NewAPI()
	assert.NoError(t, err)
	defer testApi.Close()
	testApi = testApi.NamePersister(nameP)

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
						  {"match":{"given":{"query":"freddy","boost":0.7}}},
                      	  {"match":{"given.narrow":{"query":"Fred","boost":0.6}}},
                      	  {"match":{"given.broad":{"query":"Fred","boost":0.4}}},
                      	  {"fuzzy":{"given":{"value":"fred","fuzziness":"AUTO","rewrite":"constant_score_boolean","boost":0.3}}},
                      	  {"match":{"given":{"query":"F","boost":0.2}}}
                    	]}}
					  ]}}
					]}},"from":0,"size":10}`,
		},
		{
			req: SearchRequest{
				Given:            "Fred",
				GivenFuzziness:   FuzzyNameExact | FuzzyNameVariants,
				Surname:          "Flintstone",
				SurnameFuzziness: FuzzyNameExact,
			},
			query: `{"query":{"bool":{"must":[
					  {"bool":{"must":[
						{"dis_max":{"queries":[
						  {"match":{"given":{"query":"Fred","boost":1}}},
						  {"match":{"given":{"query":"freddy","boost":0.7}}}
                    	]}},
						{"match":{"surname":{"query":"Flintstone","boost":1}}}
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
                      	  {"match":{"given.narrow":{"query":"Fred","boost":0.6}}}
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
		result, err := testApi.constructSearchQuery(ctx, &test.req)
		assert.NoError(t, err)
		//bs, _ := json.Marshal(result)
		//fmt.Printf("%d expected=%s\n", i, test.query)
		//fmt.Printf("%d actual  =%s\n", i, string(bs))
		assert.EqualValues(t, search, *result, i)
	}
}
