package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/ourrootsorg/cms-server/api"

	"github.com/gorilla/schema"
)

var decoder = schema.NewDecoder()

// Search returns search results matching a query
// @summary returns search results
// @description * Names can include wildcards (* or ?), in which case name fuzziness above Exact is ignored
// @description * Date searching is limited to passing in a single year; use fuzziness for ranges
// @description * Name fuzziness flags (OR'd together): 0: default; 1: exact; 2: variant spellings; 4: narrow sounds-like; 8: broad sounds-like; 16: fuzzy (levenshtein); 32: initials (applies only to given)
// @description * Date fuzziness: 0: default; 1: exact to this year; 2: +/- 1 year; 3: +/- 2 years; 4: +/- 5 years; 5: +/- 10 years
// @description * Places can include wildcards (* or ?) or ~word to fuzzy-match word, in which case place fuzziness above Exact is ignored
// @description * Place fuzziness flags (OR'd together): 0: default; 1: exact only; 2: include higher-level jurisdictions;
// @description * Category and collection facets: to start set categoryFacet true. If the user selects a value from the returned list, set that value as the category filter and set collectionFacet true
// @description * Date and place faceting are in a state of flux currently and may not be supported in the future depending upon user interest; do not use
// @description * Date facets: to start set century faceting to true. If the user selects a value from the returned list, set that value as the century filter and set decade faceting to true. If the user selects a decade, set that value as the decade filter
// @description * Place facets: to start, set level 1 faceting to true. If the user selects a value from the returned list, set that value as the level 1 filter and set level 2 faceting to true. Continue up to level 3
// @router /search [get]
// @tags search
// @id search
// @produce application/json
// @param given query string false "principal given and middle names"
// @param givenFuzziness query int false "principal given name fuzziness flags"
// @param surname query string false "principal surname(s)"
// @param surnameFuzziness query int false "principal surname fuzziness flags"
// @param fatherGiven query string false "father given and middle names"
// @param fatherGivenFuzziness query int false "father given name fuzziness flags"
// @param fatherSurname query string false "father surname(s)"
// @param fatherSurnameFuzziness query int false "father surname fuzziness flags"
// @param motherGiven query string false "mother given and middle names"
// @param motherGivenFuzziness query int false "mother given name fuzziness flags"
// @param motherSurname query string false "mother surname(s)"
// @param motherSurnameFuzziness query int false "mother surname fuzziness flags"
// @param spouseGiven query string false "spouse given and middle names"
// @param spouseGivenFuzziness query int false "spouse given name fuzziness flags"
// @param spouseSurname query string false "spouse surname(s)"
// @param spouseSurnameFuzziness query int false "spouse surname fuzziness flags"
// @param otherGiven query string false "other person given and middle names"
// @param otherGivenFuzziness query int false "other person given name fuzziness flags"
// @param otherSurname query string false "other person surname(s)"
// @param otherSurnameFuzziness query int false "other person surname fuzziness flags"
// @param birthDate query string false "date"
// @param birthDateFuzziness query int false "+/- year range"
// @param birthPlace query string false "place"
// @param birthPlaceFuzziness query int false "fuzziness flags"
// @param marriageDate query string false "date"
// @param marriageDateFuzziness query int false "+/- year range"
// @param marriagePlace query string false "place"
// @param marriagePlaceFuzziness query int false "fuzziness flags"
// @param residenceDate query string false "date"
// @param residenceDateFuzziness query int false "+/- year range"
// @param residencePlace query string false "place"
// @param residencePlaceFuzziness query int false "fuzziness flags"
// @param deathDate query string false "date"
// @param deathDateFuzziness query int false "+/- year range"
// @param deathPlace query string false "place"
// @param deathPlaceFuzziness query int false "fuzziness flags"
// @param anyDate query string false "date"
// @param anyDateFuzziness query int false "+/- year range"
// @param anyPlace query string false "place"
// @param anyPlaceFuzziness query int false "fuzziness flags"
// @param keywords query string false "text search on the keywords field"
// @param collectionPlace1Facet query bool false "facet on collection location level 1"
// @param collectionPlace1 query string false "filter on collection location level 1"
// @param collectionPlace2Facet query bool false "facet on collection location level 2"
// @param collectionPlace2 query string false "filter on collection location level 2"
// @param collectionPlace3Facet query bool false "facet on collection location level 3"
// @param collectionPlace3 query string false "filter on collection location level 3"
// @param categoryFacet query bool false "facet on category"
// @param category query string false "filter on category"
// @param collectionFacet query bool false "facet on collection"
// @param collection query string false "filter on collection"
// @param from query int false "starting result to return (default 0, max 1000)"
// @param size query int false "number of results to return (default 10, max 100)"
// @success 200 {array} model.SearchResult "OK"
// @failure 500 {object} api.Error "Server error"
func (app App) Search(w http.ResponseWriter, req *http.Request) {
	var searchRequest api.SearchRequest
	err := decoder.Decode(&searchRequest, req.URL.Query())
	if err != nil {
		msg := fmt.Sprintf("Bad request: %v", err.Error())
		ErrorResponse(w, http.StatusBadRequest, msg)
		return
	}

	result, errors := app.api.Search(req.Context(), &searchRequest)
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}

	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	err = enc.Encode(result)
	if err != nil {
		serverError(w, err)
		return
	}
}

// SearchByID returns detailed information about a single search result
// @summary returns a single search result
// @router /search/{id} [get]
// @tags search
// @id searchByID
// @Param id path string true "Search Result ID"
// @produce application/json
// @success 200 {object} model.SearchHit "OK"
// @failure 404 {object} api.Error "Not found"
// @failure 500 {object} api.Error "Server error"
func (app App) SearchByID(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	result, errors := app.api.SearchByID(req.Context(), vars["id"])
	if errors != nil {
		ErrorsResponse(w, errors)
		return
	}

	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", contentType)
	err := enc.Encode(result)
	if err != nil {
		serverError(w, err)
		return
	}
}
