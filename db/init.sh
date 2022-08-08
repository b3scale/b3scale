#!/usr/bin/env sh

######################################################################
# @author      : annika
# @file        : init
# @created     : Wednesday Oct 21, 2020 21:05:07 CEST
#
# @description :  Initialize the database
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


## Commandline opts: 
OPT_USAGE=0
OPT_TESTING=0
OPT_CLEAR=0

while [ $# -gt 0 ]; do
  case "$1" in
    -h) OPT_USAGE=1 ;;
    -t) OPT_TESTING=1 ;;
    -c) OPT_CLEAR=1 ;;
  esac
  shift
done

if [ $OPT_USAGE -eq 1 ]; then
    echo "Options:"
    echo "   -h     Show this helpful text"
    echo "   -c     Drop and create the database"
    echo "   -t     Make a test database"
    exit
fi

if [ $OPT_TESTING -eq 1 ]; then
    echo "++ using test database"
    DB_NAME="${DB_NAME}_test"
    export PGDATABASE=$DB_NAME
fi

if [ $OPT_CLEAR -eq 1 ]; then
    echo "++ clearing database"
    $PSQL template1 -c "DROP DATABASE $DB_NAME"
    $PSQL template1 -c "CREATE DATABASE $DB_NAME"
fi

## Apply sql scripts
$PSQL -v ON_ERROR_STOP=on < schema/0001_initial_tables.sql
$PSQL -v ON_ERROR_STOP=on < schema/0002_node_agent_api.sql

