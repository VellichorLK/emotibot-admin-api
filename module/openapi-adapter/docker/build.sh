#!/bin/bash
set -eu

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BUILDROOT=$DIR/../
REPO=docker-reg.emotibot.com.cn:55688
#Use Build root as container name since our dir name should match the binary name too.
CONTAINER=$(cd $BUILDROOT;echo "${PWD##*/}")
TAG=$(cat "$DIR/VERSION")
DOCKER_IMAGE=$REPO/$CONTAINER:$TAG

# Build docker
cmd="docker build \
  -t $DOCKER_IMAGE \
  -f $DIR/Dockerfile $BUILDROOT"
echo $cmd
eval $cmd
cmd="docker tag $DOCKER_IMAGE $REPO/$CONTAINER:latest"
echo $cmd
eval $cmd