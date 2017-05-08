#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# environment file check.
source $1
if [ "$?" -ne 0 ]; then
  echo "Erorr, can't open envfile: $1"
  echo "Usage: $0 <env file> <docker image tag>"
  echo "e.g., "
  echo " $0 dev.env 94crazy"
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
if [ ! -d "${DB_DATA_PATH}" ]; then
    mkdir -p ${DB_DATA_PATH}
fi

# check environment MONGO_DATA_PATH
if [ "${MONGO_DATA_PATH}" == "" ]; then
	MONGO_DATA_PATH="$DIR/../mongo_storage"
fi

MONGO_AUTH="--auth"
# check if mongo database persistant path exist
if [ ! -d "${MONGO_DATA_PATH}" ]; then
    mkdir -p ${MONGO_DATA_PATH}
    mkdir -p $MONGO_DATA_PATH/db
	mkdir -p $MONGO_DATA_PATH/configdb
    # if this is first time mongo deployed, do not use authentication
    MONGO_AUTH=""
fi

echo $MONGO_DATA_PATH

cmd="DB_DATA_PATH=${DB_DATA_PATH} MONGO_DATA_PATH=${MONGO_DATA_PATH}  MONGO_AUTH=${MONGO_AUTH} docker-compose -f $DIR/docker-compose.yml up --force-recreate -d"
echo $cmd
eval $cmd
