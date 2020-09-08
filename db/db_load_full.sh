#!/bin/bash
user=${1:-postgres}
password=${2:-postgres}
host=${3:-localhost}
port=${4:-5432}

wget https://s3.amazonaws.com/public.ourroots.org/place_settings.tsv
wget https://s3.amazonaws.com/public.ourroots.org/places.tsv
wget https://s3.amazonaws.com/public.ourroots.org/place_words.tsv
wget https://s3.amazonaws.com/public.ourroots.org/givenname_variants.tsv
wget https://s3.amazonaws.com/public.ourroots.org/surname_variants.tsv

./db_load_core.sh . $user $password $host $port

rm place_settings.tsv
rm places.tsv
rm place_words.tsv
rm givenname_variants.tsv
rm surname_variants.tsv
