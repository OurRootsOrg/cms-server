#!/bin/bash
psql -U postgres -f create_db.sql
migrate -database postgres://localhost:5432/ourroots?sslmode=disable -path migrations up