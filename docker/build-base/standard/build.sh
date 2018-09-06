#!/bin/bash
set -e

REPO=docker-reg.emotibot.com.cn:55688
CONTAINER=gobase
VERSION=$1
TAG=$VERSION
DOCKER_IMAGE=$REPO/$CONTAINER:$TAG

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
printf $DOCKER_IMAGE > $DIR/DOCKER_IMAGE
set -a
echo -n "Trying to pull image: $DOCKER_IMAGE...";
docker pull $DOCKER_IMAGE > /dev/null 2>&1 && echo "OK" && exit 0;
echo "FAIL";
set -e
echo "Start to build base platform image"

cmd="docker build \
        --build-arg VERSION=$VERSION \
        -t $DOCKER_IMAGE \
        -f Dockerfile ."
echo $cmd && eval $cmd && docker push $DOCKER_IMAGE;
