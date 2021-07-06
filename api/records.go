package api

import (
	"context"
	"log"

	"github.com/ourrootsorg/cms-server/model"
)

// RecordsResult is a paged Record result
type RecordsResult struct {
	Records  []model.Record `json:"records"`
	NextPage string         `json:"next_page"`
}

// RecordDetail is a record with optional labels, citation, household, and image path
type RecordDetail struct {
	model.Record
	Labels    []HeaderLabel  `json:"labels"`
	Citation  string         `json:"citation"`
	Household []model.Record `json:"household"`
	ImagePath string         `json:"imagePath"`
}
type HeaderLabel struct {
	Header string `json:"header"`
	Label  string `json:"label"`
}

// GetRecordsForPost holds the business logic around getting up to limit Records for a post
func (api API) GetRecordsForPost(ctx context.Context, postID uint32, limit int) (*RecordsResult, error) {
	// TODO: handle search criteria and paged results
	records, err := api.recordPersister.SelectRecordsForPost(ctx, postID, limit)
	if err != nil {
		return nil, NewError(err)
	}
	return &RecordsResult{Records: records}, nil
}

// GetRecordsByID holds the business logic around getting many Records
func (api API) GetRecordsByID(ctx context.Context, ids []uint32, enforceContextSocietyMatch bool) ([]model.Record, error) {
	records, err := api.recordPersister.SelectRecordsByID(ctx, ids, enforceContextSocietyMatch)
	if err != nil {
		return nil, NewError(err)
	}
	return records, nil
}

// GetRecord holds the business logic around getting a Record
func (api API) GetRecord(ctx context.Context, includeDetails bool, id uint32) (*RecordDetail, error) {
	record, err := api.recordPersister.SelectOneRecord(ctx, id)
	if err != nil {
		return nil, NewError(err)
	}
	if !includeDetails {
		return &RecordDetail{
			Record: *record,
		}, nil
	}

	// populate labels, citation, household members, and image path
	var labels []HeaderLabel
	var householdMembers []model.Record
	var imagePath string
	var citation string

	// read post and collection
	post, err := api.GetPost(ctx, record.Post)
	if err != nil {
		return nil, NewError(err)
	}
	coll, err := api.GetCollection(ctx, post.Collection)
	if err != nil {
		return nil, NewError(err)
	}
	// get labels
	for _, mapping := range coll.Mappings {
		labels = append(labels, HeaderLabel{Header: mapping.Header, Label: mapping.DbField})
	}
	// get citation
	if coll.CitationTemplate != "" {
		citation = record.GetCitation(coll.CitationTemplate)
	}
	// get household records
	if coll.HouseholdNumberHeader != "" {
		household, err := api.recordPersister.SelectOneRecordHousehold(ctx, post.ID, record.Data[coll.HouseholdNumberHeader])
		if err != nil {
			return nil, NewError(err)
		}
		memberRecords, err := api.recordPersister.SelectRecordsByID(ctx, household.Records, true)
		if err != nil {
			return nil, NewError(err)
		}
		for _, recID := range household.Records {
			for _, mbrRec := range memberRecords {
				if recID == mbrRec.ID {
					householdMembers = append(householdMembers, mbrRec)
					break
				}
			}
		}
	}
	// get image path
	if coll.ImagePathHeader != "" {
		imagePath = record.Data[coll.ImagePathHeader]
	}
	return &RecordDetail{
		Record:    *record,
		Labels:    labels,
		Citation:  citation,
		Household: householdMembers,
		ImagePath: imagePath,
	}, nil
}

// AddRecord holds the business logic around adding a Record
func (api API) AddRecord(ctx context.Context, in model.RecordIn) (*model.Record, error) {
	err := api.validate.Struct(in)
	if err != nil {
		log.Printf("[ERROR] Invalid record %v", err)
		return nil, NewError(err)
	}
	// insert
	record, e := api.recordPersister.InsertRecord(ctx, in)
	if e != nil {
		return nil, NewError(e)
	}
	//log.Printf("[DEBUG] Added record ID %d", record.ID)
	return record, nil
}

// UpdateRecord holds the business logic around updating a Record
func (api API) UpdateRecord(ctx context.Context, id uint32, in model.Record) (*model.Record, error) {
	err := api.validate.Struct(in)
	if err != nil {
		return nil, NewError(err)
	}
	record, e := api.recordPersister.UpdateRecord(ctx, id, in)
	if e != nil {
		return nil, NewError(e)
	}
	return record, nil
}

// DeleteRecord holds the business logic around deleting a Record
func (api API) DeleteRecord(ctx context.Context, id uint32) error {
	err := api.recordPersister.DeleteRecord(ctx, id)
	if err != nil {
		return NewError(err)
	}
	return nil
}

// DeleteRecordsForPost holds the business logic around deleting the Records for a Post
func (api API) DeleteRecordsForPost(ctx context.Context, postID uint32) error {
	err := api.recordPersister.DeleteRecordsForPost(ctx, postID)
	if err != nil {
		return NewError(err)
	}
	return nil
}

// GetRecordHouseholdsForPost holds the business logic around getting all Record Households for a post
func (api API) GetRecordHouseholdsForPost(ctx context.Context, postID uint32) ([]model.RecordHousehold, error) {
	recordHouseholds, err := api.recordPersister.SelectRecordHouseholdsForPost(ctx, postID)
	if err != nil {
		return nil, NewError(err)
	}
	return recordHouseholds, nil
}

// GetRecordHousehold holds the business logic around getting a Record Household
func (api API) GetRecordHousehold(ctx context.Context, postID uint32, householdID string) (*model.RecordHousehold, error) {
	recordHousehold, err := api.recordPersister.SelectOneRecordHousehold(ctx, postID, householdID)
	if err != nil {
		return nil, NewError(err)
	}
	return recordHousehold, nil
}

// AddRecordHousehold holds the business logic around adding a Record Household
func (api API) AddRecordHousehold(ctx context.Context, in model.RecordHouseholdIn) (*model.RecordHousehold, error) {
	err := api.validate.Struct(in)
	if err != nil {
		log.Printf("[ERROR] Invalid record household%v", err)
		return nil, NewError(err)
	}
	// insert
	recordHousehold, e := api.recordPersister.InsertRecordHousehold(ctx, in)
	if e != nil {
		return nil, NewError(e)
	}
	log.Printf("[DEBUG] Added record Household post=%d household=%s\n", recordHousehold.Post, recordHousehold.Household)
	return recordHousehold, nil
}

// DeleteRecordHouseholdsForPost holds the business logic around deleting the Record Households for a Post
func (api API) DeleteRecordHouseholdsForPost(ctx context.Context, postID uint32) error {
	err := api.recordPersister.DeleteRecordHouseholdsForPost(ctx, postID)
	if err != nil {
		return NewError(err)
	}
	return nil
}
