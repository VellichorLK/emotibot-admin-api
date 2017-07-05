#!/bin/bash
REPO=docker-reg.emotibot.com.cn:55688
CONTAINER=golang-webserver
TAG=20170518
DOCKER_IMAGE=$REPO/$CONTAINER:$TAG

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
EMOTIGO=${DIR#*/emotigo/}
BUILDROOT=${DIR%/emotigo/*}
GOSRCPATH="emotigo/$EMOTIGO/../"
echo $DIR
echo $GOSRCPATH
echo $BUILDROOT
echo $EMOTIGO
# Build docker
cmd="docker build \
  -t $DOCKER_IMAGE \
  --build-arg PROJECT=$GOSRCPATH \
  -f $DIR/Dockerfile $BUILDROOT"
echo $cmd
eval $cmd
