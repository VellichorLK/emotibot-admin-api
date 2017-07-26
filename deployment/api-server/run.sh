#!/bin/bash

# number of worker-voice-emotion-analysis
num_of_worker_analysis=$NUM_ANA_WORKER
if [ "$num_of_worker_analysis" == "" ]; then
	num_of_worker_analysis=5
fi
echo "num_of_worker_analysis: $num_of_worker_analysis"

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

source $1
if [ "$?" -ne 0 ]; then
  echo "Erorr, can't open envfile: $1"
  echo "Usage: $0 <env file>"
  echo "e.g., "
  echo " $0 api-sh.env [SERVICE1] [SERVICE2]..."
  exit 1
else
  envfile=$1
  echo "# Using envfile: $envfile"
fi
shift

while [ $# != 0 ]
do
    echo $1
    if [ "$1" == "worker-voice-emotion-analysis" ]; then
        service="$service "$1
        scale="--scale $1=$num_of_worker_analysis"
    fi
    shift
done

if [ "$service" == "" ]; then
    scale="--scale worker-voice-emotion-analysis=$num_of_woker_analysis"
fi
# prepare docker-compose env file
cp $envfile .env

docker-compose -f ./docker-compose.yml rm -s $service
cmd="docker-compose -f ./docker-compose.yml up --force-recreate -d $scale $service" 
echo $cmd
eval $cmd
