#!/bin/bash

REPO=docker-reg.emotibot.com.cn:55688
# The name of the container, should use the name of the repo is possible
CONTAINER=admin-api
#TAG="$(git rev-parse --short HEAD)"
LAST_RELEASE_TAG="latest"
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

ENVFILE=$1
TAG=$2

if [ "$TAG" == "" ]; then
    TAG="$LAST_RELEASE_TAG"
fi
DOCKER_IMAGE=$REPO/$CONTAINER:$TAG

if [ ${USER} == "deployer" ]; then
globalConf="
    -e TZ=Asia/Taipei \
    --restart always \
"
fi

if ! [ "$ENVFILE" == "" ]; then
    envConf="--env-file $ENVFILE"
fi

moduleConf="-p 8182:8182"

docker rm -f $CONTAINER
cmd="docker run -d --name $CONTAINER \
    $envConf \
    $globalConf \
    $moduleConf \
    $DOCKER_IMAGE
"

echo $cmd
eval $cmd
