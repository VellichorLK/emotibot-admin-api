#!/bin/bash
REPO=docker-reg.emotibot.com.cn:55688
CONTAINER=authentication-test
TAG="20170822001"
DOCKER_IMAGE=$REPO/$CONTAINER:$TAG

# Build docker
cmd="docker build \
  -t $DOCKER_IMAGE \
  -f Dockerfile ."
echo $cmd
eval $cmd
