#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

source $1
if [ "$?" -ne 0 ]; then
  echo "Erorr, can't open envfile: $1"
  echo "Usage: $0 <env file>"
  echo "e.g., "
  echo " $0 api-sh.env [SERVICE]"
  exit 1
else
  envfile=$1
  echo "# Using envfile: $envfile"
fi

service=$2

# prepare docker-compose env file
cp $envfile .env

cmd="docker-compose -f ./docker-compose.yml up --force-recreate -d $service"
echo $cmd
eval $cmd
