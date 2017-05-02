#!/bin/bash
REPO=docker-reg.emotibot.com.cn:55688
CONTAINER=goproxy-uid
TAG="$(git rev-parse --short HEAD)"
DOCKER_IMAGE=$REPO/$CONTAINER:$TAG

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BUILDROOT=$DIR/../../../
CURDIR=${PWD##*/}
GOSRCPATH=${DIR#*backend/}

echo $GOSRCPATH
echo $BUILDROOT
# Build docker
cmd="docker build \
  -t $DOCKER_IMAGE \
  --build-arg PROJECT=$GOSRCPATH \
  -f $DIR/Dockerfile $BUILDROOT"
echo $cmd
eval $cmd
