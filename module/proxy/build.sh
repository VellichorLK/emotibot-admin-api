#!/bin/bash
REPO=docker-reg.emotibot.com.cn:55688
CONTAINER=goproxy-uid
TAG=20170728006
DOCKER_IMAGE=$REPO/$CONTAINER:$TAG

# Define the environment, e.g.,
# DIR=/home/wmyao/workspaces/emotigo/module/proxy
# BUILDROOT=/home/wmyao/workspaces/emotigo/module/proxy/../../
# PROJECT=module/proxy
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BUILDROOT=$DIR/../../        #buildroot = root of emotigo repo
PROJECT=${DIR#*emotigo/}     #project = module/proxy
echo "BUILDROOT=$BUILDROOT"
echo "PROJECT=$PROJECT"

# Build docker
cmd="docker build \
  -t $DOCKER_IMAGE \
  --build-arg PROJECT=$PROJECT \
  -f $DIR/Dockerfile $BUILDROOT"
echo $cmd
eval $cmd
