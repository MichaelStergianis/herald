#!/bin/bash

createDB () {
    DB_DATABASE_EXISTS=$(psql -U postgres postgres -tAc "select 1 from pg_database where datistemplate = false and datname = '$1';")
    if [ -z $DB_DATABASE_EXISTS ]; then
	echo "Creating $1"
	createdb -U postgres -O herald $1 "The database for the herald web server"
	psql -U herald $1 -f music_schema.sql
    else
	echo "Database $1 exists"
    fi
}

cleanDB () {
    dropdb -U postgres --if-exists $1
}

DB_USER_EXISTS=$(psql -U postgres postgres -tAc "select 1 from pg_roles where rolname='herald'")

if [ -z $DB_USER_EXISTS ]; then
    echo "User herald does not exist"
    createuser herald
else
    echo "User herald exists."
fi

if [ $1 = "clean" ]; then
    echo "Cleaning databases"
    cleanDB herald
    cleanDB herald_test
fi
			       

createDB herald

createDB herald_test
