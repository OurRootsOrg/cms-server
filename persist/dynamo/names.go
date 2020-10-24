package dynamo

import (
	"bufio"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/ourrootsorg/cms-server/model"
)

const (
	givenNameVariantsType = "givenNameVariants"
	surnameVariantsType   = "surnameVariants"
)

// SelectNameVariants selects the NameVariants object if it exists or returns ErrNoRows
func (p Persister) SelectNameVariants(ctx context.Context, nameType model.NameType, name string) (*model.NameVariants, error) {
	var nv model.NameVariants
	var nt string
	switch nameType {
	case model.GivenType:
		nt = givenNameVariantsType
	case model.SurnameType:
		nt = surnameVariantsType
	default:
		return nil, model.NewError(model.ErrOther, fmt.Sprintf("Unknown name type %d", nameType))
	}

	gii := &dynamodb.GetItemInput{
		TableName: p.tableName,
		Key: map[string]*dynamodb.AttributeValue{
			pkName: {
				S: aws.String(name),
			},
			skName: {
				S: aws.String(nt),
			},
		},
	}
	gio, err := p.svc.GetItem(gii)
	if err != nil {
		log.Printf("[ERROR] Failed to get name variants. qi: %#v err: %v", gio, err)
		return nil, model.NewError(model.ErrOther, err.Error())
	}
	if gio.Item == nil {
		return nil, model.NewError(model.ErrNotFound, name)
	}
	err = dynamodbattribute.UnmarshalMap(gio.Item, &nv)
	if err != nil {
		log.Printf("[ERROR] Failed to unmarshal. qo: %#v err: %v", gio, err)
		return nil, model.NewError(model.ErrOther, err.Error())
	}
	return &nv, nil
}

// LoadNameVariantsData loads name variants data in TSV format
func (p Persister) LoadNameVariantsData(rd io.Reader, nameType model.NameType) error {
	var nameVariantsType string
	switch nameType {
	case model.GivenType:
		nameVariantsType = givenNameVariantsType
	case model.SurnameType:
		nameVariantsType = surnameVariantsType
	default:
		return model.NewError(model.ErrOther, fmt.Sprintf("Unknown name type %d", nameType))
	}

	err := p.truncateEntity(nameVariantsType)
	if err != nil {
		log.Printf("[ERROR] Failed to truncate name variants. err: %v", err)
		return model.NewError(model.ErrOther, err.Error())
	}
	r := csv.NewReader(bufio.NewReader(rd))
	r.Comma = '\t'
	r.LazyQuotes = true
	var batchCount, total int
	batch := make([]model.NameVariants, 0)
	now := time.Now().Truncate(0)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			msg := fmt.Sprintf("Error reading name variants from TSV file: %#v", err)
			log.Printf("[ERROR] " + msg)
			return errors.New(msg)
		}
		if len(record) != 2 {
			msg := fmt.Sprintf("Expected 2 fields, found %d in record(%#v)", len(record), record)
			log.Printf("[ERROR] " + msg)
			return errors.New(msg)
		}
		name := record[0]
		variants, err := decodeStringSlice(record[1], "Variants")
		if err != nil {
			return err
		}

		nv := model.NameVariants{
			Name:           name,
			Type:           nameVariantsType,
			Variants:       variants,
			InsertTime:     now,
			LastUpdateTime: now,
		}
		batch = append(batch, nv)
		batchCount++
		total++
		if batchCount >= batchSize {
			err = p.writeNameVariantsBatch(batch, total)
			if err != nil {
				return err
			}
			batchCount = 0
			batch = make([]model.NameVariants, 0)
		}
	}
	if batchCount > 0 {
		err := p.writeNameVariantsBatch(batch, total)
		if err != nil {
			return err
		}
	}
	log.Printf("[INFO] Wrote %d name variants to %s", total, nameVariantsType)
	return nil
}
func (p Persister) writeNameVariantsBatch(batch []model.NameVariants, count int) error {
	ris := map[string][]*dynamodb.WriteRequest{}
	for _, nv := range batch {
		avs, err := dynamodbattribute.MarshalMap(nv)
		if err != nil {
			msg := fmt.Sprintf("[ERROR] Failed to marshal name variant %#v: %v", nv, err)
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
		msg := fmt.Sprintf("[ERROR] Failed to write name variant batch %#v: %v", bwii, err)
		log.Printf("[ERROR] " + msg)
		return errors.New(msg)
	}
	log.Printf("[INFO] Wrote batch of %d name variants , total %d", len(batch), count)
	return nil
}
