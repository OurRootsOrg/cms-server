#!/bin/bash
LOCAL_TEST=true DYNAMODB_TABLE_NAME=test-cms FILE_PATHS=../../test_data/place_settings.tsv,../../test_data/places.tsv,../../test_data/place_words.tsv ./ddbloader
# LOCAL_TEST=true DYNAMODB_TABLE_NAME=test-cms FILE_PATH=../../test_data/givenname_variants.tsv ./ddbloader
# LOCAL_TEST=true DYNAMODB_TABLE_NAME=test-cms FILE_PATH=../../test_data/surname_variants.tsv ./ddbloader

