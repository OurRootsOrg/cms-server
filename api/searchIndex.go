package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ourrootsorg/cms-server/stddate"
	"github.com/ourrootsorg/cms-server/stdplace"

	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/ourrootsorg/cms-server/model"
	"github.com/ourrootsorg/cms-server/utils"
)

const numWorkers = 5

type GivenSurname struct {
	given   string
	surname string
}

type nameExtractor func(GivenSurname) string

var IndexRoles = map[model.Role]string{
	model.PrincipalRole:   "",
	model.FatherRole:      "f",
	model.MotherRole:      "m",
	model.SpouseRole:      "s",
	model.BrideRole:       "b",
	model.GroomRole:       "g",
	model.BrideFatherRole: "bf",
	model.BrideMotherRole: "bm",
	model.GroomFatherRole: "gf",
	model.GroomMotherRole: "gm",
	model.OtherRole:       "o",
}

var IndexRolesReversed = reverseRoleMap(IndexRoles)

func reverseRoleMap(m map[model.Role]string) map[string]model.Role {
	result := map[string]model.Role{}
	for k, v := range m {
		result[v] = k
	}
	return result
}

var RelativeRoles = map[model.Role]map[model.Relative][]model.Role{
	model.PrincipalRole: {
		model.FatherRelative: {model.FatherRole},
		model.MotherRelative: {model.MotherRole},
		model.SpouseRelative: {model.SpouseRole},
		model.OtherRelative:  {model.BrideRole, model.GroomRole, model.BrideFatherRole, model.BrideMotherRole, model.GroomFatherRole, model.GroomMotherRole, model.OtherRole},
	},
	model.FatherRole: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {model.MotherRole},
		model.OtherRelative:  {model.PrincipalRole, model.SpouseRole, model.BrideRole, model.GroomRole, model.BrideFatherRole, model.BrideMotherRole, model.GroomFatherRole, model.GroomMotherRole, model.OtherRole},
	},
	model.MotherRole: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {model.FatherRole},
		model.OtherRelative:  {model.PrincipalRole, model.SpouseRole, model.BrideRole, model.GroomRole, model.BrideFatherRole, model.BrideMotherRole, model.GroomFatherRole, model.GroomMotherRole, model.OtherRole},
	},
	model.SpouseRole: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {model.PrincipalRole},
		model.OtherRelative:  {model.FatherRole, model.MotherRole, model.BrideRole, model.GroomRole, model.BrideFatherRole, model.BrideMotherRole, model.GroomFatherRole, model.GroomMotherRole, model.OtherRole},
	},
	model.BrideRole: {
		model.FatherRelative: {model.BrideFatherRole},
		model.MotherRelative: {model.BrideMotherRole},
		model.SpouseRelative: {model.GroomRole},
		model.OtherRelative:  {model.PrincipalRole, model.FatherRole, model.MotherRole, model.SpouseRole, model.GroomFatherRole, model.GroomMotherRole, model.OtherRole},
	},
	model.GroomRole: {
		model.FatherRelative: {model.GroomFatherRole},
		model.MotherRelative: {model.GroomMotherRole},
		model.SpouseRelative: {model.BrideRole},
		model.OtherRelative:  {model.PrincipalRole, model.FatherRole, model.MotherRole, model.SpouseRole, model.BrideFatherRole, model.BrideMotherRole, model.OtherRole},
	},
	model.BrideFatherRole: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {model.BrideMotherRole},
		model.OtherRelative:  {model.PrincipalRole, model.FatherRole, model.MotherRole, model.SpouseRole, model.BrideRole, model.GroomRole, model.GroomFatherRole, model.GroomMotherRole, model.OtherRole},
	},
	model.BrideMotherRole: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {model.BrideFatherRole},
		model.OtherRelative:  {model.PrincipalRole, model.FatherRole, model.MotherRole, model.SpouseRole, model.BrideRole, model.GroomRole, model.GroomFatherRole, model.GroomMotherRole, model.OtherRole},
	},
	model.GroomFatherRole: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {model.GroomMotherRole},
		model.OtherRelative:  {model.PrincipalRole, model.FatherRole, model.MotherRole, model.SpouseRole, model.BrideRole, model.GroomRole, model.BrideFatherRole, model.BrideMotherRole, model.OtherRole},
	},
	model.GroomMotherRole: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {model.GroomFatherRole},
		model.OtherRelative:  {model.PrincipalRole, model.FatherRole, model.MotherRole, model.SpouseRole, model.BrideRole, model.GroomRole, model.BrideFatherRole, model.BrideMotherRole, model.OtherRole},
	},
	model.OtherRole: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {},
		model.OtherRelative:  {model.PrincipalRole, model.FatherRole, model.MotherRole, model.SpouseRole, model.BrideRole, model.GroomRole, model.BrideFatherRole, model.BrideMotherRole, model.GroomFatherRole, model.GroomMotherRole},
	},
}

var RelativeRelationshipsToHead = map[model.HouseholdRelToHead]map[model.Relative][]model.HouseholdRelToHead{
	model.HeadRelToHead: {
		model.FatherRelative: {model.FatherRelToHead},
		model.MotherRelative: {model.MotherRelToHead},
		model.SpouseRelative: {model.SpouseRelToHead, model.HusbandRelToHead, model.WifeRelToHead},
		model.OtherRelative:  {model.HeadRelToHead, model.ChildRelToHead, model.SonRelToHead, model.DaughterRelToHead, model.OtherRelToHead},
	},
	model.FatherRelToHead: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {model.MotherRelToHead},
		model.OtherRelative:  {model.HeadRelToHead, model.FatherRelToHead, model.SpouseRelToHead, model.HusbandRelToHead, model.WifeRelToHead, model.ChildRelToHead, model.SonRelToHead, model.DaughterRelToHead, model.OtherRelToHead},
	},
	model.MotherRelToHead: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {model.FatherRelToHead},
		model.OtherRelative:  {model.HeadRelToHead, model.MotherRelToHead, model.SpouseRelToHead, model.HusbandRelToHead, model.WifeRelToHead, model.ChildRelToHead, model.SonRelToHead, model.DaughterRelToHead, model.OtherRelToHead},
	},
	model.SpouseRelToHead: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {model.HeadRelToHead},
		model.OtherRelative:  {model.FatherRelToHead, model.MotherRelToHead, model.SpouseRelToHead, model.HusbandRelToHead, model.WifeRelToHead, model.ChildRelToHead, model.SonRelToHead, model.DaughterRelToHead, model.OtherRelToHead},
	},
	model.HusbandRelToHead: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {model.HeadRelToHead},
		model.OtherRelative:  {model.FatherRelToHead, model.MotherRelToHead, model.SpouseRelToHead, model.HusbandRelToHead, model.WifeRelToHead, model.ChildRelToHead, model.SonRelToHead, model.DaughterRelToHead, model.OtherRelToHead},
	},
	model.WifeRelToHead: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {model.HeadRelToHead},
		model.OtherRelative:  {model.FatherRelToHead, model.MotherRelToHead, model.SpouseRelToHead, model.HusbandRelToHead, model.WifeRelToHead, model.ChildRelToHead, model.SonRelToHead, model.DaughterRelToHead, model.OtherRelToHead},
	},
	model.ChildRelToHead: {
		model.FatherRelative: {model.HeadRelToHead, model.SpouseRelToHead, model.HusbandRelToHead},
		model.MotherRelative: {model.HeadRelToHead, model.SpouseRelToHead, model.WifeRelToHead},
		model.SpouseRelative: {},
		model.OtherRelative:  {model.FatherRelToHead, model.MotherRelToHead, model.ChildRelToHead, model.SonRelToHead, model.DaughterRelToHead, model.OtherRelToHead},
	},
	model.SonRelToHead: {
		model.FatherRelative: {model.HeadRelToHead, model.SpouseRelToHead, model.HusbandRelToHead},
		model.MotherRelative: {model.HeadRelToHead, model.SpouseRelToHead, model.WifeRelToHead},
		model.SpouseRelative: {},
		model.OtherRelative:  {model.FatherRelToHead, model.MotherRelToHead, model.ChildRelToHead, model.SonRelToHead, model.DaughterRelToHead, model.OtherRelToHead},
	},
	model.DaughterRelToHead: {
		model.FatherRelative: {model.HeadRelToHead, model.SpouseRelToHead, model.HusbandRelToHead},
		model.MotherRelative: {model.HeadRelToHead, model.SpouseRelToHead, model.WifeRelToHead},
		model.SpouseRelative: {},
		model.OtherRelative:  {model.FatherRelToHead, model.MotherRelToHead, model.ChildRelToHead, model.SonRelToHead, model.DaughterRelToHead, model.OtherRelToHead},
	},
	model.OtherRelToHead: {
		model.FatherRelative: {},
		model.MotherRelative: {},
		model.SpouseRelative: {},
		model.OtherRelative:  {model.HeadRelToHead, model.FatherRelToHead, model.MotherRelToHead, model.SpouseRelToHead, model.HusbandRelToHead, model.WifeRelToHead, model.ChildRelToHead, model.SonRelToHead, model.DaughterRelToHead, model.OtherRelToHead},
	},
}

// IndexPost
func (api API) IndexPost(ctx context.Context, post *model.Post) error {
	var countSuccessful uint64

	societyID, err := utils.GetSocietyIDFromContext(ctx)
	if err != nil {
		log.Printf("[ERROR] Missing society %v\n", err)
		return err
	}

	lastModified := strconv.FormatInt(time.Now().Unix()*1000, 10)

	// read collection for post
	collection, errs := api.GetCollection(ctx, post.Collection)
	if errs != nil {
		log.Printf("[ERROR] GetCollection %v\n", errs)
		return errs
	}
	// read categories for post
	categories, errs := api.GetCategoriesByID(ctx, collection.Categories)
	if errs != nil {
		log.Printf("[ERROR] GetCategory %v\n", errs)
		return errs
	}
	// read records for post
	records, errs := api.GetRecordsForPost(ctx, post.ID)
	if errs != nil {
		log.Printf("[ERROR] GetRecordsForPost %v\n", errs)
		return errs
	}

	// read record households for post
	var recordHouseholds []model.RecordHousehold
	if collection.HouseholdNumberHeader != "" {
		recordHouseholds, errs = api.GetRecordHouseholdsForPost(ctx, post.ID)
		if errs != nil {
			log.Printf("[ERROR] GetRecordHouseholdsForPost %v\n", errs)
			return errs
		}
	}

	// create the bulk indexer
	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:      "records",  // The default index name
		Client:     api.es,     // The Elasticsearch client
		NumWorkers: numWorkers, // The number of worker goroutines
	})
	if err != nil {
		log.Printf("[ERROR] Error creating the bulk indexer: %s", err)
		return err
	}
	biClosed := false
	defer func() {
		if !biClosed {
			_ = bi.Close(ctx)
		}
	}()

	householdRecordsMap := map[string][]*model.Record{}
	if collection.HouseholdNumberHeader != "" {
		householdRecordsMap = getHouseholdRecordsMap(recordHouseholds, records.Records)
	}

	for _, record := range records.Records {
		var householdRecords []*model.Record
		if collection.HouseholdNumberHeader != "" {
			householdRecords = householdRecordsMap[record.Data[collection.HouseholdNumberHeader]]
		}
		err = indexRecord(&record, householdRecords, societyID, post, collection, categories, lastModified, &countSuccessful, bi)
		if err != nil {
			log.Printf("[ERROR] Unexpected error %d: %v", record.ID, err)
			return err
		}
	}

	if err := bi.Close(ctx); err != nil {
		log.Printf("[ERROR] Unexpected error %v\n", err)
		return err
	}
	biClosed = true

	biStats := bi.Stats()
	if biStats.NumFailed > 0 {
		msg := fmt.Sprintf("[ERROR] Failed to index %d records\n", biStats.NumFailed)
		log.Printf(msg)
		return errors.New(msg)
	}

	log.Printf("[INFO] Indexed %d records\n", biStats.NumFlushed)
	return nil
}

func getHouseholdRecordsMap(recordHouseholds []model.RecordHousehold, records []model.Record) map[string][]*model.Record {
	recordsMap := map[uint32]*model.Record{}
	for ix := range records {
		recordsMap[records[ix].ID] = &records[ix]
	}
	result := map[string][]*model.Record{}
	for _, recordHousehold := range recordHouseholds {
		if recordHousehold.Household == "" {
			continue // should never happen
		}
		var records []*model.Record
		for _, recordID := range recordHousehold.Records {
			records = append(records, recordsMap[recordID])
		}
		result[recordHousehold.Household] = records
	}
	return result
}

func indexRecord(record *model.Record, householdRecords []*model.Record, societyID uint32, post *model.Post, collection *model.Collection,
	categories []model.Category, lastModified string, countSuccessful *uint64, bi esutil.BulkIndexer) error {

	for role, suffix := range IndexRoles {
		if suffix != "" {
			suffix = "_" + suffix
		}
		// get data for role
		data := getDataForRole(collection.Mappings, record, role)

		// populate the record to index
		ixRecord := map[string]interface{}{}
		ixRecord["given"] = data["given"]
		ixRecord["surname"] = data["surname"]
		if ixRecord["given"] == "" && ixRecord["surname"] == "" {
			if role == "principal" {
				log.Printf("[DEBUG] No given name or surname found for record %#v, mappings %#v, role %s",
					record, collection.Mappings, role)
			}
			continue
		}

		// get relatives' names
		for _, relative := range model.Relatives {
			names := getNames(collection.Mappings, record, RelativeRoles[role][relative])
			// include relatives' names from household
			if role == model.PrincipalRole && collection.HouseholdRelationshipHeader != "" && len(householdRecords) > 0 {
				relToHead := stdRelToHead(record.Data[collection.HouseholdRelationshipHeader])
				householdNames := getHouseholdNames(collection.HouseholdRelationshipHeader, collection.GenderHeader,
					collection.Mappings, relative, RelativeRelationshipsToHead[relToHead][relative], record.ID, householdRecords)
				if len(householdNames) > 0 {
					names = append(names, householdNames...)
				}
			}
			givens := unique(getNameParts(names, func(name GivenSurname) string { return name.given }))
			surnames := unique(getNameParts(names, func(name GivenSurname) string { return name.surname }))
			if len(givens) > 0 {
				ixRecord[string(relative)+"Given"] = strings.Join(givens, " ")
			}
			if len(surnames) > 0 {
				ixRecord[string(relative)+"Surname"] = strings.Join(surnames, " ")
			}
		}

		// get events
		for _, eventType := range model.EventTypes {
			if data[string(eventType)+"Date"] != "" {
				dates, years, valid := getDatesYears(data[string(eventType)+"Date_std"])
				if valid {
					ixRecord[string(eventType)+"DateStd"] = dates
					ixRecord[string(eventType)+"Year"] = years
				}
			}
			if data[string(eventType)+"Place"] != "" {
				placeLevels := getPlaceLevels(data[string(eventType)+"Place_std"])
				if len(placeLevels) > 0 {
					ixRecord[string(eventType)+"Place"] = data[string(eventType)+"Place"]
					ixRecord[string(eventType)+"Place1"] = placeLevels[0]
				}
				if len(placeLevels) > 1 {
					ixRecord[string(eventType)+"Place2"] = placeLevels[1]
				}
				if len(placeLevels) > 2 {
					ixRecord[string(eventType)+"Place3"] = placeLevels[2]
				}
				if len(placeLevels) > 3 {
					ixRecord[string(eventType)+"Place4"] = placeLevels[3]
				}
			}
		}

		// keywords
		ixRecord["keywords"] = data["keywords"]

		// get other data
		var catNames []string
		for _, cat := range categories {
			catNames = append(catNames, cat.Name)
		}
		ixRecord["post"] = post.ID
		ixRecord["category"] = catNames
		ixRecord["collection"] = collection.Name
		ixRecord["collectionId"] = collection.ID
		ixRecord["societyId"] = societyID
		if collection.PrivacyLevel&model.PrivacyPrivateSearch == 0 {
			ixRecord["privacy"] = Public
		}
		if collection.Location != "" {
			placeLevels := getPlaceFacets(collection.Location)
			if len(placeLevels) > 0 {
				ixRecord["collectionPlace1"] = placeLevels[0]
			}
			if len(placeLevels) > 1 {
				ixRecord["collectionPlace2"] = placeLevels[1]
			}
			if len(placeLevels) > 2 {
				ixRecord["collectionPlace3"] = placeLevels[2]
			}
		}
		ixRecord["lastModified"] = lastModified

		// add to BulkIndexer
		bs, err := json.Marshal(ixRecord)
		if err != nil {
			log.Printf("[ERROR] encoding record %d: %v", record.ID, err)
			return err
		}

		// Add an item to the BulkIndexer
		err = bi.Add(
			context.Background(),
			esutil.BulkIndexerItem{
				// Action field configures the operation to perform (index, create, delete, update)
				Action: "index",

				DocumentID: strconv.Itoa(int(record.ID)) + suffix,

				// Body is an `io.Reader` with the payload
				Body: bytes.NewReader(bs),

				// OnSuccess is called for each successful operation
				OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
					atomic.AddUint64(countSuccessful, 1)
				},

				// OnFailure is called for each failed operation
				OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
					if err != nil {
						log.Printf("[ERROR]: %s", err)
					} else {
						log.Printf("[ERROR]: %s: %s", res.Error.Type, res.Error.Reason)
					}
				},
			},
		)
		if err != nil {
			log.Printf("[ERROR] indexing record %d: %v\n", record.ID, err)
		}
	}

	return nil
}

func getNameParts(names []GivenSurname, extractor nameExtractor) []string {
	var parts []string
	for _, name := range names {
		part := extractor(name)
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}

func getYears(dateParts, dateRange []string) []int {
	years := []int{}
	switch {
	case len(dateParts) == 1:
		year, err := strconv.Atoi(dateParts[0][:4])
		if err != nil {
			break
		}
		years = append(years, year)
	case len(dateParts) == 2 && len(dateRange) == 1:
		firstYear, err := strconv.Atoi(dateParts[0][:4])
		if err != nil {
			break
		}
		years = append(years, firstYear)
		secondYear, err := strconv.Atoi(dateParts[1][:4])
		if err != nil {
			break
		}
		if secondYear != firstYear {
			years = append(years, secondYear)
		}
	case len(dateParts) == 2 && len(dateRange) == 2:
		startYear, err := strconv.Atoi(dateRange[0][:4])
		if err != nil {
			break
		}
		endYear, err := strconv.Atoi(dateRange[1][:4])
		if err != nil {
			break
		}
		for y := startYear; y <= endYear; y++ {
			years = append(years, y)
		}
	}
	return years

}

func getDatesYears(encodedDate string) ([]int, []int, bool) {
	if encodedDate == "" {
		return nil, nil, false
	}
	// parse encoded date
	dateParts := strings.Split(encodedDate, ",")
	var dateRange []string
	if len(dateParts) == 2 {
		dateRange = strings.Split(dateParts[1], "-")
	}

	// get dates
	dates := []int{}
	for i := 0; i < len(dateParts); i++ {
		ymd, err := strconv.Atoi(dateParts[i])
		if err != nil {
			return nil, nil, false
		}
		dates = append(dates, ymd)
		// get just one date for range
		if len(dateRange) == 2 {
			break
		}
	}

	// get years
	years := getYears(dateParts, dateRange)

	return dates, years, true
}

func getPlaceLevels(stdPlace string) []string {
	var stdLevels []string
	if stdPlace == "" {
		return stdLevels
	}
	levels := splitPlace(stdPlace)
	var std string
	for i := len(levels) - 1; i >= 0; i-- {
		std += strings.TrimSpace(levels[i])
		if i > 0 {
			std += ","
		}
		stdLevels = append(stdLevels, std)
	}
	return stdLevels
}

func getPlaceFacets(stdPlace string) []string {
	var stdLevels []string
	if stdPlace == "" {
		return stdLevels
	}
	levels := splitPlace(stdPlace)
	for i := len(levels) - 1; i >= 0; i-- {
		stdLevels = append(stdLevels, strings.TrimSpace(levels[i]))
	}
	return stdLevels
}

func unique(arr []string) []string {
	var result []string
OUTER:
	for _, s := range arr {
		for _, res := range result {
			if res == s {
				continue OUTER
			}
		}
		result = append(result, s)
	}
	return result
}

func getDataForRole(mappings []model.CollectionMapping, record *model.Record, role model.Role) map[string]string {
	data := map[string]string{}

	for _, mapping := range mappings {
		// get marriage data for spouse too
		if record.Data[mapping.Header] != "" &&
			(mapping.IxRole == string(role) || (isSpouseRole(mapping.IxRole, role) && isMarriageField(mapping.IxField))) {
			data[mapping.IxField] = record.Data[mapping.Header]
			if strings.HasSuffix(mapping.IxField, "Date") {
				data[mapping.IxField+stddate.StdSuffix] = record.Data[mapping.Header+stddate.StdSuffix]
			} else if strings.HasSuffix(mapping.IxField, "Place") {
				data[mapping.IxField+stdplace.StdSuffix] = record.Data[mapping.Header+stdplace.StdSuffix]
			}
		}
	}
	return data
}

func isMarriageField(field string) bool {
	return field == "marriageDate" || field == "marriagePlace"
}

func isSpouseRole(role1 string, role2 model.Role) bool {
	switch role1 {
	case string(model.PrincipalRole):
		return role2 == model.SpouseRole
	case string(model.SpouseRole):
		return role2 == model.PrincipalRole
	case string(model.FatherRole):
		return role2 == model.MotherRole
	case string(model.MotherRole):
		return role2 == model.FatherRole
	case string(model.BrideRole):
		return role2 == model.GroomRole
	case string(model.GroomRole):
		return role2 == model.BrideRole
	case string(model.BrideFatherRole):
		return role2 == model.BrideMotherRole
	case string(model.BrideMotherRole):
		return role2 == model.BrideFatherRole
	case string(model.GroomFatherRole):
		return role2 == model.GroomMotherRole
	case string(model.GroomMotherRole):
		return role2 == model.GroomFatherRole
	}
	return false
}

func getNames(mappings []model.CollectionMapping, record *model.Record, roles []model.Role) []GivenSurname {
	names := []GivenSurname{}

	for _, role := range roles {
		names = append(names, getNamesForRole(role, mappings, record)...)
	}
	return names
}

func getNamesForRole(role model.Role, mappings []model.CollectionMapping, record *model.Record) []GivenSurname {
	var names []GivenSurname

	var givens []string
	var surnames []string
	for _, mapping := range mappings {
		if mapping.IxRole == string(role) {
			if mapping.IxField == "given" && record.Data[mapping.Header] != "" {
				givens = append(givens, record.Data[mapping.Header])
			}
			if mapping.IxField == "surname" && record.Data[mapping.Header] != "" {
				surnames = append(surnames, record.Data[mapping.Header])
			}
		}
	}
	if len(givens) > 0 || len(surnames) > 0 {
		givens = unique(givens)
		surnames = unique(surnames)
		names = append(names, GivenSurname{
			given:   strings.Join(givens, " "),
			surname: strings.Join(surnames, " "),
		})
	}
	return names
}

func getHouseholdNames(relToHeadHeader, genderHeader string, mappings []model.CollectionMapping, relative model.Relative,
	relsToHead []model.HouseholdRelToHead, recordID uint32, householdRecords []*model.Record) []GivenSurname {

	names := []GivenSurname{}

	// for each record in household that is not this record
	for _, record := range householdRecords {
		if record.ID == recordID {
			continue
		}

		// if the record's relationship to head is in relsToHead, get the names
		recordRelToHead := stdRelToHead(record.Data[relToHeadHeader])
		found := false
		for _, relToHead := range relsToHead {
			if recordRelToHead == relToHead {
				found = true
				break
			}
		}
		if !found {
			continue
		}

		// if the relative is father or mother and the relationship is head or spouse, make sure the gender doesn't disagree
		if (relative == model.FatherRelative || relative == model.MotherRelative) &&
			(recordRelToHead == model.HeadRelToHead || recordRelToHead == model.SpouseRelToHead) {
			recordGender := stdGender(record.Data[genderHeader])
			if (relative == model.FatherRelative && recordGender == model.GenderFemale) ||
				(relative == model.MotherRelative && recordGender == model.GenderMale) {
				continue
			}
		}

		// get names
		names = append(names, getNamesForRole(model.PrincipalRole, mappings, record)...)
	}
	return names
}

func stdRelToHead(relToHead string) model.HouseholdRelToHead {
	relToHead = strings.ToLower(relToHead)
	for _, stdRelToHead := range model.HouseholdRelsToHead {
		if relToHead == string(stdRelToHead) {
			return stdRelToHead
		}
	}
	return model.OtherRelToHead
}

func stdGender(gender string) model.Gender {
	gender = strings.ToLower(gender)
	if strings.HasPrefix(gender, "f") {
		return model.GenderFemale
	}
	if strings.HasPrefix(gender, "m") {
		return model.GenderMale
	}
	return model.GenderOther
}

var placeRegexp = regexp.MustCompile("\\s*,\\s*")

func splitPlace(place string) []string {
	return placeRegexp.Split(place, -1)
}
