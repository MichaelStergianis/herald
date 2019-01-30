#!/bin/bash

DB_USER_EXISTS=$(psql -U postgres postgres -tAc "select 1 from pg_roles where rolname='herald'")

if [ $DB_USER_EXISTS == "1" ]; then
    echo "User herald exists:" $DB_USER_EXISTS "."
fi

if [ -z $DB_USER_EXISTS ]; then
    echo "User herald does not exist"
fi
			       

DB_DATABASE_EXISTS=$(psql -U postgres postgres -tAc "select 1 from pg_database where datistemplate = false and datname = 'herald_db';")
if [ -z $DB_DATABASE_EXISTS ]; then
    echo "Creating herald_db"
    createdb -U postgres -O herald herald_db "The database for the herald web server"
fi
if [ $DB_DATABASE_EXISTS ]; then
    echo "Database herald_db exists"
fi
