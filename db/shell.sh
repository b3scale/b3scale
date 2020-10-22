#!/usr/bin/env sh

######################################################################
# @author      : annika
# @file        : shell 
# @created     : Wednesday Oct 21, 2020 21:15:17 CEST
#
# @description : Get a psql shell with the defaults 
######################################################################

if [ -z $PSQL ]; then
    PSQL="psql"
fi

if [ -z $DB_HOST ]; then
    DB_HOST="localhost"
fi

if [ -z $DB_PORT ]; then
    DB_PORT="5432"
fi

if [ -z $DB_NAME ]; then
    DB_NAME="b3scale"
fi

if [ -z $DB_USER ]; then
    DB_USER="postgres"
fi

if [ -z $DB_PASSWORD ]; then
    DB_PASSWORD="postgres"
fi


## Setup postgres env
export PGHOST=$DB_HOST
export PGPORT=$DB_PORT
export PGDATABASE=$DB_NAME
export PGUSER=$DB_USER
export PGPASSWORD=$DB_PASSWORD

# Start postgres client
$PSQL
