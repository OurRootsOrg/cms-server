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
// @description * Names can include wildcards (* or ?). In that case name fuzziness is ignored
// @description * Date and place searching are not yet implemented. They will be implemented in August.
// @description * Name fuzziness flags (OR'd together): 0: exact only; 1: alternate spellings; 2: narrow sounds-like; 4: broad sounds-like; 8: fuzzy (levenshtein); 16: initials (applies only to given)
// @description * Date fuzziness: +/- number of years to generate a year range
// @description * Place fuzziness flags (OR'd together): 0: exact only; 1: include higher-level jurisdictions; 2: include nearby places
// @description * Date facets: to start set century faceting to true. If ht user selects a value from the returned list, set that value as the century filter and set decade faceting to true. If the user selects a decade, set that value as the decade filter
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
// @param fatherFuzziness query int false "father surname fuzziness flags"
// @param motherGiven query string false "mother given and middle names"
// @param motherGivenFuzziness query int false "mother given name fuzziness flags"
// @param motherSurname query string false "mother surname(s)"
// @param motherFuzziness query int false "mother surname fuzziness flags"
// @param spouseGiven query string false "spouse given and middle names"
// @param spouseGivenFuzziness query int false "spouse given name fuzziness flags"
// @param spouseSurname query string false "spouse surname(s)"
// @param spouseFuzziness query int false "spouse surname fuzziness flags"
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
// @param birthCenturyFacet query bool false "facet on century"
// @param birthCentury query string false "filter on century"
// @param birthDecadeFacet query bool false "facet on decade"
// @param birthDecade query string false "filter on decade"
// @param birthPlace1Facet query bool false "facet on place level 1"
// @param birthPlace1 query string false "filter on place level 1"
// @param birthPlace2Facet query bool false "facet on place level 2"
// @param birthPlace2 query string false "filter on place level 2"
// @param birthPlace3Facet query bool false "facet on place level 3"
// @param birthPlace3 query string false "filter on place level 3"
// @param marriageCenturyFacet query bool false "facet on century"
// @param marriageCentury query string false "filter on century"
// @param marriageDecadeFacet query bool false "facet on decade"
// @param marriageDecade query string false "filter on decade"
// @param marriagePlace1Facet query bool false "facet on place level 1"
// @param marriagePlace1 query string false "filter on place level 1"
// @param marriagePlace2Facet query bool false "facet on place level 2"
// @param marriagePlace2 query string false "filter on place level 2"
// @param marriagePlace3Facet query bool false "facet on place level 3"
// @param marriagePlace3 query string false "filter on place level 3"
// @param residenceCenturyFacet query bool false "facet on century"
// @param residenceCentury query string false "filter on century"
// @param residenceDecadeFacet query bool false "facet on decade"
// @param residenceDecade query string false "filter on decade"
// @param residencePlace1Facet query bool false "facet on place level 1"
// @param residencePlace1 query string false "filter on place level 1"
// @param residencePlace2Facet query bool false "facet on place level 2"
// @param residencePlace2 query string false "filter on place level 2"
// @param residencePlace3Facet query bool false "facet on place level 3"
// @param residencePlace3 query string false "filter on place level 3"
// @param deathCenturyFacet query bool false "facet on century"
// @param deathCentury query string false "filter on century"
// @param deathDecadeFacet query bool false "facet on decade"
// @param deathDecade query string false "filter on decade"
// @param deathPlace1Facet query bool false "facet on place level 1"
// @param deathPlace1 query string false "filter on place level 1"
// @param deathPlace2Facet query bool false "facet on place level 2"
// @param deathPlace2 query string false "filter on place level 2"
// @param deathPlace3Facet query bool false "facet on place level 3"
// @param deathPlace3 query string false "filter on place level 3"
// @param otherCenturyFacet query bool false "facet on century"
// @param otherCentury query string false "filter on century"
// @param otherDecadeFacet query bool false "facet on decade"
// @param otherDecade query string false "filter on decade"
// @param otherPlace1Facet query bool false "facet on place level 1"
// @param otherPlace1 query string false "filter on place level 1"
// @param otherPlace2Facet query bool false "facet on place level 2"
// @param otherPlace2 query string false "filter on place level 2"
// @param otherPlace3Facet query bool false "facet on place level 3"
// @param otherPlace3 query string false "filter on place level 3"
// @param categoryFacet query bool false "facet on category"
// @param category query string false "filter on category"
// @param collectionFacet query bool false "facet on collection"
// @param collection query string false "filter on collection"
// @success 200 {array} model.SearchResult "OK"
// @failure 500 {object} api.Error "Server error"
// TODO need to specify possible query parameters
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
