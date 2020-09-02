package model

type SearchResult struct {
	Hits     []SearchHit            `json:"hits"`
	Total    int                    `json:"total"`
	MaxScore float64                `json:"maxScore"`
	Facets   map[string]SearchFacet `json:"facets"`
}
type SearchHit struct {
	ID                 string       `json:"id"`
	Score              float64      `json:"score"`
	Person             SearchPerson `json:"person,omitempty"`
	Record             SearchRecord `json:"record,omitempty"` // only returned on search by id
	CollectionID       uint32       `json:"collection"`
	CollectionName     string       `json:"collectionName"`
	CollectionLocation string       `json:"collectionLocation,omitempty"` // only returned on search by id
	Citation           string       `json:"citation,omitempty"`           // only returned on search by id
	PostID             uint32       `json:"post,omitempty"`               // only returned on search by id
	ImagePath          string       `json:"imagePath,omitempty"`          // only returned on search by id
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
type SearchFacet struct {
	ErrorUpperBound int                 `json:"errorUpperBound"`
	OtherDocCount   int                 `json:"otherDocCount"`
	Buckets         []SearchFacetBucket `json:"buckets"`
}
type SearchFacetBucket struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}
type SearchRecord []SearchLabelValue
type SearchLabelValue struct {
	Label string `json:"label"`
	Value string `json:"value"`
}
