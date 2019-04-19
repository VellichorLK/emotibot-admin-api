#!/bin/bash
REPO=docker-reg.emotibot.com.cn:55688
CONTAINER=vipshop-fake-auth-base
# TAG="$(git rev-parse --short HEAD)"
TAG="2019032801-rc1"
DOCKER_IMAGE=$REPO/$CONTAINER:$TAG

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
GOSRCPATH="$(cd "$DIR/../" && pwd )"
MODULE=${GOSRCPATH##/*/}
BUILDROOT=$DIR/../../

# Build docker
cmd="docker build \
  -t $DOCKER_IMAGE \
  --build-arg PROJECT=$MODULE \
  -f $DIR/Dockerfile-base $BUILDROOT"
echo $cmd
eval $cmd
