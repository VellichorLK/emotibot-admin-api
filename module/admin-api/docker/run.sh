#!/bin/bash

REPO=docker-reg.emotibot.com.cn:55688
# The name of the container, should use the name of the repo is possible
CONTAINER=admin-api
#TAG="$(git rev-parse --short HEAD)"
LAST_RELEASE_TAG="latest"

TAG=$2
if [ "$TAG" == "" ]; then
    TAG="$LAST_RELEASE_TAG"
fi
DOCKER_IMAGE=$REPO/$CONTAINER:$TAG

if [ ${USER} == "deployer" ]; then
globalConf="
    -v /etc/localtime:/etc/localtime \
    --restart always \
"
fi
moduleConf="
    -p 8181:8181
    --env-file $1
"

docker rm -f $CONTAINER
cmd="docker run -d --name $CONTAINER \
    $globalConf \
    $moduleConf \
    $DOCKER_IMAGE \
"

echo $cmd
eval $cmd
