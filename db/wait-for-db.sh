#!/usr/bin/env bash
# Use this script to wait for postgres to really be ready
port=${1:-5432} # can override port from command line
until PGPASSWORD=postgres psql -U postgres -h localhost -p $port -c '\q'; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep 1
done