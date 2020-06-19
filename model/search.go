package model

type SearchResult struct {
	Hits     []SearchHit `json:"hits"`
	Total    int         `json:"total"`
	MaxScore float64     `json:"maxScore"`
}
type SearchHit struct {
	ID             string            `json:"id"`
	Score          float64           `json:"score"`
	Person         SearchPerson      `json:"person,omitempty"`
	Record         map[string]string `json:"record,omitempty"` // only returned on search by id
	CollectionName string            `json:"collectionName"`
	CollectionID   uint32            `json:"collection"`
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
