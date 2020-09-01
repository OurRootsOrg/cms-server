#!/bin/bash
port=${1:-5432} # can override port from command line
PGPASSWORD=postgres psql -U postgres -h localhost -p $port -d cms -c "\copy place(id, name, full_name, alt_names, types, located_in_id, also_located_in_ids, level, country_id, latitude, longitude, count) FROM 'test_data/places.tsv'"
PGPASSWORD=postgres psql -U postgres -h localhost -p $port -d cms -c "\copy place_word(word, ids) FROM 'test_data/place_words.tsv'"
PGPASSWORD=postgres psql -U postgres -h localhost -p $port -d cms -c "\copy place_settings(id, body) FROM 'test_data/place_settings.tsv'"
PGPASSWORD=postgres psql -U postgres -h localhost -p $port -d cms -c "\copy givenname_variants(name, variants) FROM 'test_data/givenname_variants.tsv'"
PGPASSWORD=postgres psql -U postgres -h localhost -p $port -d cms -c "\copy surname_variants(name, variants) FROM 'test_data/surname_variants.tsv'"
