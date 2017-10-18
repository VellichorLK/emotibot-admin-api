#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# DB_DATA_PATH=/home/deployer/persistant_storage ./run.sh

# environment file check.
#source $1
#if [ "$?" -ne 0 ]; then
#  echo "Erorr, can't open envfile: $1"
#  echo "Usage: $0 <env file>"
#  echo "e.g., "
#  echo " $0 dev.env"
#  exit 1
#else
#  envfile=$1
#  echo "# Using envfile: $envfile"
#fi
#
## prepare docker-compose env file
#cp $envfile .env

# check environment DB_DATA_PATH
if [ "${DB_DATA_PATH}" == "" ]; then
    DB_DATA_PATH="$DIR/../data"
fi

# check if database persistant path exist
if [ ! -d "${DB_DATA_PATH}/rabbitmq" ]; then
    mkdir -p "${DB_DATA_PATH}/rabbitmq"
fi


cmd="RABBITMQ_DATA_PATH=${DB_DATA_PATH}/rabbitmq docker-compose -f $DIR/docker-compose.yml up --force-recreate -d"
echo $cmd
eval $cmd
