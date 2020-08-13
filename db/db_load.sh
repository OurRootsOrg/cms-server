#!/bin/bash
port=${1:-5432} # can override port from command line
PGPASSWORD=postgres psql -U postgres -h localhost -p $port -d cms -c "\copy place(id, name, alt_names, types, located_in_id, also_located_in_ids, level, country_id, latitude, longitude, sources, count) FROM 'test_data/places.tsv'"
PGPASSWORD=postgres psql -U postgres -h localhost -p $port -d cms -c "\copy place_word(word, ids) FROM 'test_data/place_words.tsv'"
PGPASSWORD=postgres psql -U postgres -h localhost -p $port -d cms -c "\copy place_metadata(id, body) FROM 'test_data/place_metadata.tsv'"
