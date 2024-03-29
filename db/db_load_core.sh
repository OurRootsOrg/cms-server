#!/bin/bash
# can override from command line
datadir=${1:-test_data}
user=${2:-postgres}
password=${3:-postgres}
host=${4:-localhost}
port=${5:-5432}
PGPASSWORD=$password psql -U $user -h $host -p $port -d cms -c "truncate record_household, record, post, collection_category, collection, category, invitation, society_user, cms_user, society restart identity"
PGPASSWORD=$password psql -U $user -h $host -p $port -d cms -c "truncate place, place_settings, place_word, givenname_variants, surname_variants"
PGPASSWORD=$password psql -U $user -h $host -p $port -d cms -c "\copy place_settings(id, body) FROM '$datadir/place_settings.tsv'"
PGPASSWORD=$password psql -U $user -h $host -p $port -d cms -c "\copy place(id, name, full_name, alt_names, types, located_in_id, also_located_in_ids, level, country_id, latitude, longitude, count) FROM '$datadir/places.tsv'"
PGPASSWORD=$password psql -U $user -h $host -p $port -d cms -c "\copy place_word(word, ids) FROM '$datadir/place_words.tsv'"
PGPASSWORD=$password psql -U $user -h $host -p $port -d cms -c "\copy givenname_variants(name, variants) FROM '$datadir/givenname_variants.tsv'"
PGPASSWORD=$password psql -U $user -h $host -p $port -d cms -c "\copy surname_variants(name, variants) FROM '$datadir/surname_variants.tsv'"
