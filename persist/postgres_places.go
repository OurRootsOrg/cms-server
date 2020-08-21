package persist

import (
	"context"
	"regexp"
	"strings"

	"github.com/lib/pq"

	"github.com/ourrootsorg/cms-server/model"
)

const PlaceSettingsID = 1

// SelectPlaceSettings selects the PlaceSettings object if it exists or returns ErrNoRows
func (p PostgresPersister) SelectPlaceSettings(ctx context.Context) (*model.PlaceSettings, error) {
	var placeSettings model.PlaceSettings
	err := p.db.QueryRowContext(ctx, "SELECT body, insert_time, last_update_time FROM place_settings WHERE id=$1", PlaceSettingsID).Scan(
		&placeSettings.PlaceSettingsBody,
		&placeSettings.InsertTime,
		&placeSettings.LastUpdateTime,
	)
	id := uint32(PlaceSettingsID)
	return &placeSettings, translateError(err, &id, nil, "")
}

// SelectPlace selects the Place object if it exists or returns ErrNoRows
func (p PostgresPersister) SelectPlace(ctx context.Context, id uint32) (*model.Place, error) {
	var place model.Place
	err := p.db.QueryRowContext(ctx, "SELECT id, name, full_name, alt_names, types, located_in_id, also_located_in_ids, level, country_id, latitude, longitude, count, insert_time, last_update_time "+
		"FROM place WHERE id=$1", id).Scan(
		&place.ID,
		&place.Name,
		&place.FullName,
		&place.AltNames,
		&place.Types,
		&place.LocatedInID,
		&place.AlsoLocatedInIDs,
		&place.Level,
		&place.CountryID,
		&place.Latitude,
		&place.Longitude,
		&place.Count,
		&place.InsertTime,
		&place.LastUpdateTime,
	)
	return &place, translateError(err, &id, nil, "")
}

// SelectPlacesByID selects multiple Place objects by ID
func (p PostgresPersister) SelectPlacesByID(ctx context.Context, ids []uint32) ([]model.Place, error) {
	places := make([]model.Place, 0)
	if len(ids) == 0 {
		return places, nil
	}
	rows, err := p.db.QueryContext(ctx, "SELECT id, name, full_name, alt_names, types, located_in_id, also_located_in_ids, level, country_id, latitude, longitude, count, insert_time, last_update_time "+
		"FROM place WHERE id = ANY($1)", pq.Array(ids))
	if err != nil {
		return nil, translateError(err, nil, nil, "")
	}
	defer rows.Close()
	for rows.Next() {
		var place model.Place
		err := rows.Scan(
			&place.ID,
			&place.Name,
			&place.FullName,
			&place.AltNames,
			&place.Types,
			&place.LocatedInID,
			&place.AlsoLocatedInIDs,
			&place.Level,
			&place.CountryID,
			&place.Latitude,
			&place.Longitude,
			&place.Count,
			&place.InsertTime,
			&place.LastUpdateTime,
		)
		if err != nil {
			return nil, translateError(err, nil, nil, "")
		}
		places = append(places, place)
	}
	return places, nil
}

var placeRegexp = regexp.MustCompile("\\s*,\\s*")

func (p PostgresPersister) SelectPlacesByFullNamePrefix(ctx context.Context, prefix string, count int) ([]model.Place, error) {
	places := make([]model.Place, 0)
	if prefix == "" {
		return places, nil
	}
	search := strings.Join(placeRegexp.Split(prefix, -1), "%")
	if !strings.HasSuffix(search, "%") {
		search += "%"
	}
	rows, err := p.db.QueryContext(ctx, "SELECT id, name, full_name, alt_names, types, located_in_id, also_located_in_ids, level, country_id, latitude, longitude, count, insert_time, last_update_time "+
		"FROM place WHERE full_name ilike $1 ORDER BY count DESC LIMIT $2", search, count)
	if err != nil {
		return nil, translateError(err, nil, nil, "")
	}
	defer rows.Close()
	for rows.Next() {
		var place model.Place
		err := rows.Scan(
			&place.ID,
			&place.Name,
			&place.FullName,
			&place.AltNames,
			&place.Types,
			&place.LocatedInID,
			&place.AlsoLocatedInIDs,
			&place.Level,
			&place.CountryID,
			&place.Latitude,
			&place.Longitude,
			&place.Count,
			&place.InsertTime,
			&place.LastUpdateTime,
		)
		if err != nil {
			return nil, translateError(err, nil, nil, "")
		}
		places = append(places, place)
	}
	return places, nil
}

// SelectPlaceWord selects the PlaceWord object if it exists or returns ErrNoRows
func (p PostgresPersister) SelectPlaceWord(ctx context.Context, word string) (*model.PlaceWord, error) {
	var placeWord model.PlaceWord
	err := p.db.QueryRowContext(ctx, "SELECT word, ids, insert_time, last_update_time FROM place_word WHERE word=$1", word).Scan(
		&placeWord.Word,
		&placeWord.IDs,
		&placeWord.InsertTime,
		&placeWord.LastUpdateTime,
	)
	return &placeWord, translateError(err, nil, nil, "")
}

// SelectPlaceWordsByID selects multiple PlaceWord objects by word
func (p PostgresPersister) SelectPlaceWordsByWord(ctx context.Context, words []string) ([]model.PlaceWord, error) {
	placeWords := make([]model.PlaceWord, 0)
	if len(words) == 0 {
		return placeWords, nil
	}
	rows, err := p.db.QueryContext(ctx, "SELECT word, ids, insert_time, last_update_time FROM place_word WHERE word = ANY($1)", pq.Array(words))
	if err != nil {
		return nil, translateError(err, nil, nil, "")
	}
	defer rows.Close()
	for rows.Next() {
		var placeWord model.PlaceWord
		err := rows.Scan(
			&placeWord.Word,
			&placeWord.IDs,
			&placeWord.InsertTime,
			&placeWord.LastUpdateTime,
		)
		if err != nil {
			return nil, translateError(err, nil, nil, "")
		}
		placeWords = append(placeWords, placeWord)
	}
	return placeWords, nil
}
