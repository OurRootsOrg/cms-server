#!/bin/bash
set -e
ENVIRONMENT_NAME="${ENVIRONMENT_NAME:?}"
AWS_REGION="${AWS_REGION:?}"
NORMAL_THROUGHPUT="${VAR1:-5}"
LOAD_THROUGHPUT="${VAR1:-500}"

DYNAMODB_TABLE_NAME="${ENVIRONMENT_NAME}-cms" FILE_URLS=https://s3.amazonaws.com/public.ourroots.org/place_settings.tsv,https://s3.amazonaws.com/public.ourroots.org/places.tsv,https://s3.amazonaws.com/public.ourroots.org/place_words.tsv,https://s3.amazonaws.com/public.ourroots.org/givenname_variants.tsv,https://s3.amazonaws.com/public.ourroots.org/surname_variants.tsv ./ddbloader
echo NORMAL_THROUGHPUT $NORMAL_THROUGHPUT
# Only add name variants
# DYNAMODB_TABLE_NAME="${ENVIRONMENT_NAME}-cms" FILE_URLS=https://s3.amazonaws.com/public.ourroots.org/givenname_variants.tsv,https://s3.amazonaws.com/public.ourroots.org/surname_variants.tsv ./ddbloader
