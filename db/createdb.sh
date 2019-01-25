#!/bin/bash

DB_USER_EXISTS=$(psql postgres -tAc "select 1 from pg_roles where rolname='herald'")
DB_DATABASE_EXISTS=$(psql postgres -tAc "select 1 from pg_database where datistemplate = false and datname = 'herald_db';")

DB_TABLE=$(psql postgres -tAc "select 1 from pg_tables")
