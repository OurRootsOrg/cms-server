#!/bin/bash
# Use this script to wait for elasticsearch to really be ready
port=${1:-9200} # can override port from command line
set -e
until $(curl -sSf -XGET "http://localhost:${port}/_cluster/health?wait_for_status=yellow" > /dev/null); do
    sleep 2
done
echo "Elasticsearch is up on ${port}"
