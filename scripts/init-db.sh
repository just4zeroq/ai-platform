#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE user_svc;
    CREATE DATABASE asset_svc;
    CREATE DATABASE market_svc;
EOSQL
