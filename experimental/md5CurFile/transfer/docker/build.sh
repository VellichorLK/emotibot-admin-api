#!/bin/bash
REPO=docker-reg.emotibot.com.cn:55688
CONTAINER=mediamigrator
TAG=2018032800
DOCKER_IMAGE=$REPO/$CONTAINER:$TAG

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

GOSRCPATH="$(cd "$DIR/../" && pwd )"
MODULE=${GOSRCPATH##/*/}
BUILDROOT=$DIR/../../

echo $DIR
echo $GOSRCPATH
echo $MODULE
echo $BUILDROOT
# Build docker
cmd="docker build \
  -t $DOCKER_IMAGE \
  --build-arg Module=$MODULE \
  -f $DIR/Dockerfile $BUILDROOT"
echo $cmd
eval $cmd
