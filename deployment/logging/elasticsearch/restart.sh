#!/bin/bash
ELASTICSEARCH_CONTAINER="elasticsearch"

docker rm -f $ELASTICSEARCH_CONTAINER
cmd="docker run -d --name $ELASTICSEARCH_CONTAINER  -p 9200:9200 -p 9300:9300 elasticsearch"
echo $cmd
eval $cmd
