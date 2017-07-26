#!/bin/bash
REPO=docker-reg.emotibot.com.cn:55688
CONTAINER=authentication
# TAG="$(git rev-parse --short HEAD)"
TAG="20170725003"
DOCKER_IMAGE=$REPO/$CONTAINER:$TAG

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
EMOTIGO=${DIR#*/emotigo/}
BUILDROOT=${DIR%/emotigo/*}
GOSRCPATH="emotigo/$EMOTIGO/../"
echo $DIR

# BUILDROOT=${DIR%/emotigo/*}
# CURDIR=${PWD##*/}
# GOPREFIX=${DIR#*emotigo/}
# GOSRCPATH="emotigo/$GOPREFIX"
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