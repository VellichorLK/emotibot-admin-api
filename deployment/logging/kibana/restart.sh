#!/bin/bash
KIBANA_CONTAINER="kibana"
PWD=$( pwd )

docker rm -f $KIBANA_CONTAINER
cmd="docker run --name $KIBANA_CONTAINER -v $PWD/kibana.yml:/kibana.yml -d -p 5601:5601 kibana -c /kibana.yml"
echo $cmd
eval $cmd
