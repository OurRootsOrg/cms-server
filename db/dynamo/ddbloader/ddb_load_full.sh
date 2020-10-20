#!/bin/bash
set -e
ENVIRONMENT_NAME="${ENVIRONMENT_NAME:?}"
AWS_REGION="${AWS_REGION:?}"

DYNAMODB_TABLE_NAME="${ENVIRONMENT_NAME}-cms" FILE_URLS=https://s3.amazonaws.com/public.ourroots.org/place_settings.tsv,https://s3.amazonaws.com/public.ourroots.org/places.tsv,https://s3.amazonaws.com/public.ourroots.org/place_words.tsv ./ddbloader
# https://s3.amazonaws.com/public.ourroots.org/givenname_variants.tsv
# https://s3.amazonaws.com/public.ourroots.org/surname_variants.tsv
