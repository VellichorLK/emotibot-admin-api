#!/bin/bash

REPO=docker-reg.emotibot.com.cn:55688
# The name of the container, should use the name of the repo is possible
CONTAINER=vipshop-auth-adapter
#TAG="$(git rev-parse --short HEAD)"
TAG="20171129001"
DOCKER_IMAGE=$REPO/$CONTAINER:$TAG

globalConf="
    -v /etc/localtime:/etc/localtime \
    --restart always \
"
moduleConf="
    -p 8786:8786
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
