#!/bin/bash

createDB () {
    DB_DATABASE_EXISTS=$(psql -U postgres postgres -tAc "select 1 from pg_database where datistemplate = false and datname = '$1';")
    if [ -z $DB_DATABASE_EXISTS ]; then
	echo "Creating $1"
	createdb -U postgres -O warbler $1 "The database for the warbler web server"
	psql -U warbler $1 -f music_schema.sql
    else
	echo "Database $1 exists"
    fi
}

cleanDB () {
    dropdb -U postgres --if-exists $1
}

DB_USER_EXISTS=$(psql -U postgres postgres -tAc "select 1 from pg_roles where rolname='warbler'")

if [ -z $DB_USER_EXISTS ]; then
    echo "User warbler does not exist"
    createuser -U postgres warbler
else
    echo "User warbler exists."
fi

if [ "$1" = "clean" ]; then
    echo "Cleaning databases"
    cleanDB warbler
    cleanDB warbler_test
fi
			       

createDB warbler
createDB warbler_test
