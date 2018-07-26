#!/bin/bash

DB_USER=$(psql postgres -tAc "select 1 from pg_roles where rolname='herald'")
DB_TABLE=$(psql postgres -tAc "select 1 from pg_tables")
