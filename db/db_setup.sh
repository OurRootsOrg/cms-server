#!/bin/bash
port=${1:-5432} # can override port from command line
PGPASSWORD=postgres psql -U postgres -h localhost -p $port -f create_db.sql
migrate -database "postgres://postgres:postgres@localhost:$port/ourroots?sslmode=disable" -path migrations up
