#!/bin/bash
port=${1:-5432}
./db_load_core.sh test_data postgres postgres localhost $port