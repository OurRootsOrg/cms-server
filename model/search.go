package model

type Role string

const (
	PrincipalRole   Role = "principal"
	FatherRole           = "father"
	MotherRole           = "mother"
	SpouseRole           = "spouse"
	BrideRole            = "bride"
	GroomRole            = "groom"
	BrideFatherRole      = "brideFather"
	BrideMotherRole      = "brideMother"
	GroomFatherRole      = "groomFather"
	GroomMotherRole      = "groomMother"
	OtherRole            = "other"
)

// supported relationships to head: head, father, mother, spouse, husband, wife, child, son, daughter
// everything else is handled as other
type HouseholdRelToHead string

const (
	HeadRelToHead     HouseholdRelToHead = "head"
	FatherRelToHead                      = "father"
	MotherRelToHead                      = "mother"
	SpouseRelToHead                      = "spouse"
	HusbandRelToHead                     = "husband"
	WifeRelToHead                        = "wife"
	ChildRelToHead                       = "child"
	SonRelToHead                         = "son"
	DaughterRelToHead                    = "daughter"
	OtherRelToHead                       = "other"
)

var HouseholdRelsToHead = []HouseholdRelToHead{
	HeadRelToHead,
	FatherRelToHead,
	MotherRelToHead,
	SpouseRelToHead,
	HusbandRelToHead,
	WifeRelToHead,
	ChildRelToHead,
	SonRelToHead,
	DaughterRelToHead,
	OtherRelToHead,
}

// supported genders: male (anything that starts with an m), female (anything that starts with f)
// everything else is handled as other
type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale        = "female"
	GenderOther         = "other"
)

type EventType string

const (
	BirthEvent     EventType = "birth"
	MarriageEvent            = "marriage"
	ResidenceEvent           = "residence"
	DeathEvent               = "death"
	OtherEvent               = "other"
)

var EventTypes = []EventType{BirthEvent, MarriageEvent, ResidenceEvent, DeathEvent, OtherEvent}

type Relative string

const (
	FatherRelative Relative = "father"
	MotherRelative          = "mother"
	SpouseRelative          = "spouse"
	OtherRelative           = "other"
)

var Relatives = []Relative{FatherRelative, MotherRelative, SpouseRelative, OtherRelative}

type SearchResult struct {
	Hits     []SearchHit            `json:"hits"`
	Total    int                    `json:"total"`
	MaxScore float64                `json:"maxScore"`
	Facets   map[string]SearchFacet `json:"facets"`
}
type SearchHit struct {
	ID                 string         `json:"id"`
	Score              float64        `json:"score"`
	Person             SearchPerson   `json:"person,omitempty"`
	Record             SearchRecord   `json:"record,omitempty"` // only returned on search by id
	CollectionID       uint32         `json:"collection"`
	CollectionName     string         `json:"collectionName"`
	CollectionLocation string         `json:"collectionLocation,omitempty"` // only returned on search by id
	Citation           string         `json:"citation,omitempty"`           // only returned on search by id
	PostID             uint32         `json:"post,omitempty"`               // only returned on search by id
	ImagePath          string         `json:"imagePath,omitempty"`          // only returned on search by id
	Household          []SearchRecord `json:"household,omitempty"`          // only returned on search by id
}
type SearchPerson struct {
	Name          string               `json:"name"`
	Role          Role                 `json:"role"`
	Events        []SearchEvent        `json:"events,omitempty"`
	Relationships []SearchRelationship `json:"relationships,omitempty"`
}
type SearchEvent struct {
	Type  EventType `json:"type"`
	Date  string    `json:"date,omitempty"`
	Place string    `json:"place,omitempty"`
}
type SearchRelationship struct {
	Type Relative `json:"type"`
	Name string   `json:"name,omitempty"`
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
