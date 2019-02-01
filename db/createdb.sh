#!/bin/bash

createDB () {
    DB_DATABASE_EXISTS=$(psql -U postgres postgres -tAc "select 1 from pg_database where datistemplate = false and datname = '$1';")
    if [ -z $DB_DATABASE_EXISTS ]; then
	echo "Creating $1"
	createdb -U postgres -O herald $1 "The database for the herald web server"
    else
	echo "Database $1 exists"
    fi
}

DB_USER_EXISTS=$(psql -U postgres postgres -tAc "select 1 from pg_roles where rolname='herald'")

if [ -z $DB_USER_EXISTS ]; then
    echo "User herald does not exist"
    createuser herald
else
    echo "User herald exists."
fi
			       

createDB herald

createDB herald_test
