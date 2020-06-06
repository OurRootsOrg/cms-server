package model

import "fmt"

// SearchIDFormat is the format for Search IDs
const SearchIDFormat = "/search/%s"

type SearchResult struct {
	Hits  []SearchHit `json:"hits"`
	Total int         `json:"total"`
}
type SearchHit struct {
	ID             string            `json:"id"`
	Person         SearchPerson      `json:"person,omitempty"`
	Record         map[string]string `json:"record,omitempty"` // only returned on search by id
	CollectionName string            `json:"collectionName"`
	CollectionID   string            `json:"collection"`
}
type SearchPerson struct {
	Name          string               `json:"name"`
	Role          string               `json:"role"`
	Events        []SearchEvent        `json:"events,omitempty"`
	Relationships []SearchRelationship `json:"relationships,omitempty"`
}
type SearchEvent struct {
	Type  string `json:"type"`
	Date  string `json:"date,omitempty"`
	Place string `json:"place,omitempty"`
}
type SearchRelationship struct {
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

// MakeRecordID builds a Record ID string from a string ID
func MakeSearchID(id string) string {
	return pathPrefix + fmt.Sprintf(SearchIDFormat, id)
}
