#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# environment file check.
source $1
if [ "$?" -ne 0 ]; then
  echo "Erorr, can't open envfile: $1"
  echo "Usage: $0 <env file>"
  echo "e.g., "
  echo " $0 dev.env"
  exit 1
else
  envfile=$1
  echo "# Using envfile: $envfile"
fi

# prepare docker-compose env file
cp $envfile .env

# check environment DB_DATA_PATH
if [ "${DB_DATA_PATH}" == "" ]; then
    DB_DATA_PATH="$DIR/../data"
fi

# check if database persistant path exist
if [ ! -d "${DB_DATA_PATH}/mysql" ]; then
    mkdir -p "${DB_DATA_PATH}/mysql"
fi

# check if database persistant path exist
if [ ! -d "${DB_DATA_PATH}/mongo" ]; then
    mkdir -p "${DB_DATA_PATH}/mongo"
fi


cmd="MYSQL_DATA_PATH=${DB_DATA_PATH}/mysql MONGO_DATA_PATH=${DB_DATA_PATH}/mongo docker-compose -f $DIR/docker-compose.yml up --force-recreate -d"
echo $cmd
eval $cmd
