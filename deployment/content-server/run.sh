#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

source $1
if [ "$?" -ne 0 ]; then
  echo "Erorr, can't open envfile: $1"
  echo "Usage: $0 <env file>"
  echo "e.g., "
  echo " $0 content-sh.env [SERVICE1] [SERVICE2]..."
  exit 1
else
  envfile=$1
  echo "# Using envfile: $envfile"
fi
shift

if [ "$envfile" == "test.env" ]; then
    mkdir -p /tmp/persistant_storage
else
    mkdir -p /home/deployer/persistant_storage
fi

while [ $# != 0 ]
do
    echo $1
    service="$service "$1
    shift
done

# prepare docker-compose env file
cp $envfile .env

docker-compose -f ./docker-compose.yml rm -s $service

depends="--no-deps"
if [ "$service" == "" ]; then 
    depends=""
fi
cmd="docker-compose -f ./docker-compose.yml up --force-recreate --remove-orphans $depends -d $service" 
echo $cmd
eval $cmd
