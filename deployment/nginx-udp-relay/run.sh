#!/bin/sh

DOCKER_IMAGE=nginx:1-alpine
DOCKER_CONTAINER=nginx-udp-relay

#cmd="docker pull ${DOCKER_IMAGE}"
#echo $cmd
#eval $cmd

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cmd="docker rm -f ${DOCKER_CONTAINER}; docker run -d --name=${DOCKER_CONTAINER} -p 8125:8125/udp -v $DIR/nginx.conf:/etc/nginx/nginx.conf ${DOCKER_IMAGE}"
echo $cmd
eval $cmd
