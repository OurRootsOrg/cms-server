#!/bin/bash
NORMAL_THROUGHPUT="${VAR1:-5}"
LOAD_THROUGHPUT="${VAR1:-500}"

LOCAL_TEST=true DYNAMODB_TABLE_NAME=test-cms NORMAL_THROUGHPUT="${NORMAL_THROUGHPUT}" LOAD_THROUGHPUT="${LOAD_THROUGHPUT}" FILE_PATHS=../../test_data/place_settings.tsv,../../test_data/places.tsv,../../test_data/place_words.tsv,../../test_data/givenname_variants.tsv,../../test_data/surname_variants.tsv ./ddbloader

