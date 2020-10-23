package dynamo

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/ourrootsorg/cms-server/model"
)

const (
	placeSettingsType = "placeSettings"
	placeType         = "place"
	placeWordType     = "placeWord"
	batchSize         = 20
)

// SelectPlaceSettings selects the PlaceSettings object if it exists or returns ErrNoRows
func (p Persister) SelectPlaceSettings(ctx context.Context) (*model.PlaceSettings, error) {
	var ps model.PlaceSettings
	gii := &dynamodb.GetItemInput{
		TableName: p.tableName,
		Key: map[string]*dynamodb.AttributeValue{
			pkName: {
				S: aws.String(placeSettingsType),
			},
			skName: {
				S: aws.String(placeSettingsType),
			},
		},
	}
	gio, err := p.svc.GetItem(gii)
	if err != nil {
		log.Printf("[ERROR] Failed to get place settings. gii: %#v err: %v", gio, err)
		return nil, model.NewError(model.ErrOther, err.Error())
	}
	if gio.Item == nil {
		return nil, model.NewError(model.ErrNotFound, placeSettingsType)
	}
	err = dynamodbattribute.UnmarshalMap(gio.Item, &ps)
	if err != nil {
		log.Printf("[ERROR] Failed to unmarshal. qo: %#v err: %v", gio, err)
		return nil, model.NewError(model.ErrOther, err.Error())
	}
	return &ps, nil
}

// SelectPlace selects a Place object by ID
func (p Persister) SelectPlace(ctx context.Context, id uint32) (*model.Place, error) {
	ids := strconv.FormatInt(int64(id), 10)
	qi := &dynamodb.QueryInput{
		TableName:              p.tableName,
		KeyConditionExpression: aws.String(pkName + " = :pk and begins_with(" + skName + ", :sk)"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":pk": {
				S: aws.String(ids),
			},
			":sk": {
				S: aws.String(placeType + "#"),
			},
		},
	}
	places := make([]model.Place, 0)
	qo, err := p.svc.Query(qi)
	if err != nil {
		log.Printf("[ERROR] Failed to get places. qi: %#v err: %v", qi, err)
		return nil, model.NewError(model.ErrOther, err.Error())
	}
	err = dynamodbattribute.UnmarshalListOfMaps(qo.Items, &places)
	if err != nil {
		log.Printf("[ERROR] Failed to unmarshal places. qo: %#v err: %v", qo, err)
		return nil, model.NewError(model.ErrOther, err.Error())
	}
	if len(places) > 1 {
		log.Printf("[ERROR] Unexpectedly found more than one place. qo: %#v err: %v", qo, err)
		return nil, fmt.Errorf("Unexpectedly found more than one place for id %d", id)
	} else if len(places) == 0 {
		return nil, model.NewError(model.ErrNotFound, ids)
	}
	return &places[0], nil
}

// SelectPlacesByID selects multiple Place objects by ID
func (p Persister) SelectPlacesByID(ctx context.Context, ids []uint32) ([]model.Place, error) {
	// We can't do a query to select multiple Places, so just call SelectPlace in a loop
	var places []model.Place
	for _, id := range ids {
		p, err := p.SelectPlace(ctx, id)
		if err != nil {
			e, ok := err.(model.Error)
			if ok && e.Code == model.ErrNotFound {
				continue
			}
			return nil, err
		}
		places = append(places, *p)
	}
	return places, nil
}

// SelectPlaceWord selects a PlaceWord object by word
func (p Persister) SelectPlaceWord(ctx context.Context, word string) (*model.PlaceWord, error) {
	var placeWord model.PlaceWord
	gii := &dynamodb.GetItemInput{
		TableName: p.tableName,
		Key: map[string]*dynamodb.AttributeValue{
			pkName: {
				S: aws.String(placeWordType + "#" + word),
			},
			skName: {
				S: aws.String(placeWordType),
			},
		},
	}
	gio, err := p.svc.GetItem(gii)
	if err != nil {
		log.Printf("[ERROR] Failed to get place. qi: %#v err: %v", gio, err)
		return nil, model.NewError(model.ErrOther, err.Error())
	}
	if gio.Item == nil {
		return nil, model.NewError(model.ErrNotFound, word)
	}
	err = dynamodbattribute.UnmarshalMap(gio.Item, &placeWord)
	if err != nil {
		log.Printf("[ERROR] Failed to unmarshal. qo: %#v err: %v", gio, err)
		return nil, model.NewError(model.ErrOther, err.Error())
	}
	placeWord.Word = strings.TrimPrefix(placeWord.Pk, placeWordType+"#")
	return &placeWord, nil
}

// SelectPlaceWordsByWord selects multiple PlaceWord objects by word
func (p Persister) SelectPlaceWordsByWord(ctx context.Context, words []string) ([]model.PlaceWord, error) {
	placeWords := make([]model.PlaceWord, 0)
	if len(words) == 0 {
		return placeWords, nil
	}
	keys := make([]map[string]*dynamodb.AttributeValue, len(words))
	for i, word := range words {
		keys[i] = map[string]*dynamodb.AttributeValue{
			pkName: {
				S: aws.String(placeWordType + "#" + word),
			},
			skName: {
				S: aws.String(placeWordType),
			},
		}
	}
	bgii := &dynamodb.BatchGetItemInput{
		RequestItems: map[string]*dynamodb.KeysAndAttributes{
			*p.tableName: {
				Keys: keys,
			},
		},
	}
	bgio, err := p.svc.BatchGetItem(bgii)
	if err != nil {
		log.Printf("[ERROR] Failed to get places. bgii: %#v err: %v", bgii, err)
		return nil, model.NewError(model.ErrOther, err.Error())
	}
	err = dynamodbattribute.UnmarshalListOfMaps(bgio.Responses[*p.tableName], &placeWords)
	if err != nil {
		log.Printf("[ERROR] Failed to unmarshal. bgio: %#v err: %v", bgio, err)
		return nil, model.NewError(model.ErrOther, err.Error())
	}
	for _, p := range placeWords {
		p.Word = strings.TrimPrefix(p.Pk, placeWordType+"#")
	}
	return placeWords, nil
}

var placeRegexp = regexp.MustCompile("\\s*,\\s*")

// SelectPlacesByFullNamePrefix selects multiple Place objects by a prefix
func (p Persister) SelectPlacesByFullNamePrefix(ctx context.Context, prefix string, count int) ([]model.Place, error) {
	places := make([]model.Place, 0)
	if len(prefix) < 2 {
		return places, nil
	}
	matches := placeRegexp.Split(regexp.QuoteMeta(prefix), -1)
	if len(matches) == 0 {
		return places, nil
	}
	search := strings.Join(matches, ".*")
	if !strings.HasSuffix(search, ".*") {
		search += ".*"
	}
	search = "(?i)^" + search + "$"
	searchRegex, err := regexp.Compile(search)
	if err != nil {
		// This shouldn't happen, right?
		return nil, model.NewError(model.ErrOther, err.Error())
	}
	if len(matches[0]) < 2 {
		return places, nil
	}
	log.Printf("[DEBUG] search='%s', matches: %#v", search, matches)
	qi := &dynamodb.QueryInput{
		TableName:              p.tableName,
		IndexName:              aws.String(gsiName),
		KeyConditionExpression: aws.String(skName + " = :sk and begins_with(" + gsiSkName + ", :gsiSk)"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":sk": {
				S: aws.String(placeType + "#" + prefix2(strings.ToLower(prefix))),
			},
			":gsiSk": {
				S: aws.String(strings.ToLower(matches[0])),
			},
		},
	}

	for {
		batch := make([]model.Place, 0)
		qo, err := p.svc.Query(qi)
		if err != nil {
			log.Printf("[ERROR] Failed to get places. qi: %#v err: %v", qi, err)
			return nil, model.NewError(model.ErrOther, err.Error())
		}

		err = dynamodbattribute.UnmarshalListOfMaps(qo.Items, &places)
		if err != nil {
			log.Printf("[ERROR] Failed to unmarshal places. qo: %#v err: %v", qo, err)
			return nil, model.NewError(model.ErrOther, err.Error())
		}
		places = append(places, batch...)
		if qo.LastEvaluatedKey == nil {
			break
		}
		qi.ExclusiveStartKey = qo.LastEvaluatedKey
	}

	filteredPlaces := make([]model.Place, 0)
	for _, p := range places {
		if searchRegex.MatchString(p.FullName) {
			filteredPlaces = append(filteredPlaces, p)
		}
	}
	// sort by Count descending
	sort.Slice(filteredPlaces, func(i, j int) bool {
		return filteredPlaces[i].Count > filteredPlaces[j].Count
	})
	if count > len(filteredPlaces) {
		count = len(filteredPlaces)
	}
	return filteredPlaces[0:count], nil
}

// LoadPlaceSettingsData loads place settings data in TSV format
func (p Persister) LoadPlaceSettingsData(rd io.Reader) error {
	err := p.truncateEntity(placeSettingsType)
	if err != nil {
		log.Printf("[ERROR] Failed to truncate place settings. err: %v", err)
		return model.NewError(model.ErrOther, err.Error())
	}
	r := csv.NewReader(rd)
	r.Comma = '\t'
	r.LazyQuotes = true
	// Only one record, two fields. We care about the second field which is JSON
	record, err := r.Read()
	if err == io.EOF {
		msg := "No data found while reading place settings from TSV file"
		log.Printf("[ERROR] " + msg)
		return errors.New(msg)
	}
	if err != nil {
		msg := fmt.Sprintf("Error reading place settings from TSV file: %#v", err)
		log.Printf("[ERROR] " + msg)
		return errors.New(msg)
	}
	if len(record) != 2 {
		msg := fmt.Sprintf("Expected 2 fields, found %d while reading place settings", len(record))
		log.Printf("[ERROR] " + msg)
		return errors.New(msg)
	}
	dec := json.NewDecoder(strings.NewReader(record[1]))
	var ps model.PlaceSettings
	err = dec.Decode(&ps)
	if err != nil {
		msg := fmt.Sprintf("Error decoding JSON place settings: %#v", err)
		log.Printf("[ERROR] " + msg)
		return errors.New(msg)
	}
	ps.Pk = placeSettingsType
	ps.Sk = placeSettingsType
	// ps.AltSort = placeSettingsType
	now := time.Now().Truncate(0)
	ps.InsertTime = now
	ps.LastUpdateTime = now
	avs, err := dynamodbattribute.MarshalMap(ps)
	if err != nil {
		msg := fmt.Sprintf("[ERROR] Failed to marshal ps %#v: %v", ps, err)
		log.Printf("[ERROR] " + msg)
		return errors.New(msg)
	}
	pii := &dynamodb.PutItemInput{
		TableName:           p.tableName,
		Item:                avs,
		ConditionExpression: aws.String("attribute_not_exists(pk)"), // Either an insert or last_update_time must match
	}
	_, err = p.svc.PutItem(pii)
	if err != nil {
		msg := fmt.Sprintf("Failed to update ps %#v. pii: %#v err: %v", ps, pii, err)
		log.Printf("[ERROR] " + msg)
		return errors.New(msg)
	}
	return nil
}

// LoadPlaceData loads Place data in TSV format
func (p Persister) LoadPlaceData(rd io.Reader) error {
	// truncateEntity() won't work for Place, because the SK (GSI PK) isn't fixed, it's "place#" + first two characters of word.
	// So we'd have to enumerate all possible two character prefixes, i.e. 36^2.

	// err := p.truncateEntity(placeType)
	// if err != nil {
	// 	log.Printf("[ERROR] Failed to truncate places. err: %v", err)
	// 	return model.NewError(model.ErrOther, err.Error())
	// }
	r := csv.NewReader(bufio.NewReader(rd))
	r.Comma = '\t'
	r.LazyQuotes = true
	var batchCount, total int
	batch := make([]model.Place, 0)
	now := time.Now().Truncate(0)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			msg := fmt.Sprintf("Error reading places from TSV file: %#v", err)
			log.Printf("[ERROR] " + msg)
			return errors.New(msg)
		}
		if len(record) != 12 {
			msg := fmt.Sprintf("Expected 12 fields, found %d in record(%#v)", len(record), record)
			log.Printf("[ERROR] " + msg)
			return errors.New(msg)
		}
		id, err := decodeUint32(record[0], "ID")
		if err != nil {
			return err
		}
		altNames, err := decodeStringSlice(record[3], "AltNames")
		if err != nil {
			return err
		}
		types, err := decodeStringSlice(record[4], "Types")
		if err != nil {
			return err
		}
		locatedInID, err := decodeUint32(record[5], "LocatedInID")
		if err != nil {
			return err
		}
		alsoLocatedInIDs, err := decodeUint32Slice(record[6], "AlsoLocatedInIDs")
		if err != nil {
			return err
		}
		level, err := decodeUint32(record[7], "Level")
		if err != nil {
			return err
		}
		countryID, err := decodeUint32(record[8], "CountryID")
		if err != nil {
			return err
		}
		latitude, err := decodeFloat32(record[9], "Latitude")
		if err != nil {
			return err
		}
		longitude, err := decodeFloat32(record[10], "Longitude")
		if err != nil {
			return err
		}
		count, err := decodeUint32(record[11], "Count")
		if err != nil {
			return err
		}

		place := model.Place{
			ID:               id,
			Type:             placeType + "#" + prefix2(strings.ToLower(record[2])),
			AltSort:          strings.ToLower(record[2]),
			Name:             record[1],
			FullName:         record[2],
			AltNames:         altNames,
			Types:            types,
			LocatedInID:      locatedInID,
			AlsoLocatedInIDs: alsoLocatedInIDs,
			Level:            int(level),
			CountryID:        countryID,
			Latitude:         latitude,
			Longitude:        longitude,
			Count:            int(count),
			InsertTime:       now,
			LastUpdateTime:   now,
		}
		log.Printf("[DEBUG] place: %#v", place)
		batch = append(batch, place)
		batchCount++
		total++
		if batchCount >= batchSize {
			err = p.writePlaceBatch(batch, total)
			if err != nil {
				return err
			}
			batchCount = 0
			batch = make([]model.Place, 0)
		}
	}
	if batchCount > 0 {
		err := p.writePlaceBatch(batch, total)
		if err != nil {
			return err
		}
	}
	log.Printf("[INFO] Wrote %d places", total)
	return nil
}

func (p Persister) writePlaceBatch(batch []model.Place, count int) error {
	ris := map[string][]*dynamodb.WriteRequest{}
	for _, place := range batch {
		avs, err := dynamodbattribute.MarshalMap(place)
		if err != nil {
			msg := fmt.Sprintf("[ERROR] Failed to marshal place %#v: %v", place, err)
			log.Printf("[ERROR] " + msg)
			return errors.New(msg)
		}
		ris[*p.tableName] = append(ris[*p.tableName],
			&dynamodb.WriteRequest{
				PutRequest: &dynamodb.PutRequest{
					Item: avs,
				},
			},
		)
	}
	bwii := &dynamodb.BatchWriteItemInput{
		RequestItems: ris,
	}
	var bwio *dynamodb.BatchWriteItemOutput
	var err error
	for bwio == nil || (len(bwio.UnprocessedItems) > 0 && err == nil) {
		if bwio != nil && len(bwio.UnprocessedItems) > 0 {
			bwii.RequestItems = bwio.UnprocessedItems
		}
		bwio, err = p.svc.BatchWriteItem(bwii)
	}
	if err != nil {
		msg := fmt.Sprintf("[ERROR] Failed to write place batch %#v: %v", bwii, err)
		log.Printf("[ERROR] " + msg)
		return errors.New(msg)
	}
	log.Printf("[INFO] Wrote batch of %d places, total %d", len(batch), count)
	return nil
}

func decodeUint32(in string, name string) (uint32, error) {
	id, err := strconv.ParseUint(in, 10, 32)
	if err != nil {
		msg := fmt.Sprintf("Unable to decode %s uint32 (%s): %v", name, in, err)
		log.Printf("[ERROR] " + msg)
		return 0, errors.New(msg)
	}
	return uint32(id), nil
}
func decodeFloat32(in string, name string) (float32, error) {
	val, err := strconv.ParseFloat(in, 32)
	if err != nil {
		msg := fmt.Sprintf("Unable to decode %s float32 (%s): %v", name, in, err)
		log.Printf("[ERROR] " + msg)
		return 0, errors.New(msg)
	}
	return float32(val), nil
}
func decodeStringSlice(in string, name string) ([]string, error) {
	dec := json.NewDecoder(strings.NewReader(in))
	ss := make([]string, 0)
	err := dec.Decode(&ss)
	if err != nil {
		msg := fmt.Sprintf("Unable to decode %s string slice (%s): %v", name, in, err)
		log.Printf("[ERROR] " + msg)
		return nil, errors.New(msg)
	}
	return ss, nil
}
func decodeUint32Slice(in string, name string) ([]uint32, error) {
	dec := json.NewDecoder(strings.NewReader(in))
	uis := make([]uint32, 0)
	err := dec.Decode(&uis)
	if err != nil {
		msg := fmt.Sprintf("Unable to decode %s uint32 slice (%s): %v", name, in, err)
		log.Printf("[ERROR] " + msg)
		return nil, errors.New(msg)
	}
	return uis, nil
}

// LoadPlaceWordData loads PlaceWord data in TSV format
func (p Persister) LoadPlaceWordData(rd io.Reader) error {
	err := p.truncateEntity(placeWordType)
	if err != nil {
		log.Printf("[ERROR] Failed to truncate place words. err: %v", err)
		return model.NewError(model.ErrOther, err.Error())
	}
	r := csv.NewReader(bufio.NewReader(rd))
	r.Comma = '\t'
	r.LazyQuotes = true
	var batchCount, total int
	batch := make([]model.PlaceWord, 0)
	now := time.Now().Truncate(0)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			msg := fmt.Sprintf("Error reading place words from TSV file: %#v", err)
			log.Printf("[ERROR] " + msg)
			return errors.New(msg)
		}
		if len(record) != 2 {
			msg := fmt.Sprintf("Expected 2 fields, found %d in record(%#v)", len(record), record)
			log.Printf("[ERROR] " + msg)
			return errors.New(msg)
		}
		word := record[0]
		ids, err := decodeUint32Slice(record[1], "IDs")
		if err != nil {
			return err
		}

		placeWord := model.PlaceWord{
			Pk:             placeWordType + "#" + word,
			Type:           placeWordType,
			Word:           word,
			IDs:            ids,
			InsertTime:     now,
			LastUpdateTime: now,
		}
		batch = append(batch, placeWord)
		batchCount++
		total++
		if batchCount >= batchSize {
			err = p.writePlaceWordBatch(batch, total)
			if err != nil {
				return err
			}
			batchCount = 0
			batch = make([]model.PlaceWord, 0)
		}
	}
	if batchCount > 0 {
		err := p.writePlaceWordBatch(batch, total)
		if err != nil {
			return err
		}
	}
	log.Printf("[INFO] Wrote %d place words", total)
	return nil
}
func (p Persister) writePlaceWordBatch(batch []model.PlaceWord, count int) error {
	ris := map[string][]*dynamodb.WriteRequest{}
	for _, placeWord := range batch {
		avs, err := dynamodbattribute.MarshalMap(placeWord)
		if err != nil {
			msg := fmt.Sprintf("[ERROR] Failed to marshal place word %#v: %v", placeWord, err)
			log.Printf("[ERROR] " + msg)
			return errors.New(msg)
		}
		ris[*p.tableName] = append(ris[*p.tableName],
			&dynamodb.WriteRequest{
				PutRequest: &dynamodb.PutRequest{
					Item: avs,
				},
			},
		)
	}
	bwii := &dynamodb.BatchWriteItemInput{
		RequestItems: ris,
	}
	var bwio *dynamodb.BatchWriteItemOutput
	var err error
	for bwio == nil || (len(bwio.UnprocessedItems) > 0 && err == nil) {
		if bwio != nil && len(bwio.UnprocessedItems) > 0 {
			bwii.RequestItems = bwio.UnprocessedItems
		}
		bwio, err = p.svc.BatchWriteItem(bwii)
	}
	if err != nil {
		msg := fmt.Sprintf("[ERROR] Failed to write place word batch %#v: %v", bwii, err)
		log.Printf("[ERROR] " + msg)
		return errors.New(msg)
	}
	log.Printf("[INFO] Wrote batch of %d place words , total %d", len(batch), count)
	return nil
}

// prefix2 returns the first two characters of word
func prefix2(word string) string {
	prefix := word
	if len(prefix) > 2 {
		prefix = string([]rune(prefix)[0:2])
	}
	return prefix
}
