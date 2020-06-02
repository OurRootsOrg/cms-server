#!/bin/bash
port=${1:-9200} # can override port from command line
curl -X PUT "http://localhost:${port}/records" -H 'Content-Type: application/json' -d @elasticsearch_schema.json
