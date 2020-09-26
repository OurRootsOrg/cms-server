#!/bin/bash
set -e
ENVIRONMENT_NAME="${ENVIRONMENT_NAME:?}"
AWS_REGION="${AWS_REGION:?}"

DYNAMODB_TABLE_NAME="${ENVIRONMENT_NAME}-cms"
FILE_PATH=../../test_data/place_settings.tsv ./ddbloader
FILE_PATH=../../test_data/places.tsv ./ddbloader
FILE_PATH=../../test_data/place_words.tsv ./ddbloader
# FILE_PATH=../../test_data/givenname_variants.tsv ./ddbloader
# FILE_PATH=../../test_data/surname_variants.tsv ./ddbloader

