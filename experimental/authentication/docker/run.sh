#!/bin/bash

REPO=docker-reg.emotibot.com.cn:55688
# The name of the container, should use the name of the repo is possible
CONTAINER=authentication
#TAG="$(git rev-parse --short HEAD)"
TAG="2017103001_vipshop"
DOCKER_IMAGE=$REPO/$CONTAINER:$TAG

globalConf="
    -v /etc/localtime:/etc/localtime \
    --restart always \
"
moduleConf="
    -p 8088:8088
"

docker rm -f $CONTAINER
cmd="docker run -d --name $CONTAINER \
    $globalConf \
    $moduleConf \
    $DOCKER_IMAGE \
"

echo $cmd
eval $cmd
