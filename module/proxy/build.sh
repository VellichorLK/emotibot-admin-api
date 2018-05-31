#!/bin/bash
REPO=docker-reg.emotibot.com.cn:55688
CONTAINER=goproxy-uid
# Stable version before 2018/06/05
#TAG=20180321001
TAG=20180606001
DOCKER_IMAGE=$REPO/$CONTAINER:$TAG

# Define the environment, e.g.,
# DIR=/home/wmyao/workspaces/emotigo/module/proxy
# BUILDROOT=/home/wmyao/workspaces/emotigo/module/proxy/../../
# PROJECT=module/proxy
BUILDROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
echo "BUILDROOT=$BUILDROOT"


# Build docker
cmd="docker build \
  -t $DOCKER_IMAGE \
  -f $BUILDROOT/Dockerfile $BUILDROOT"
echo $cmd
eval $cmd
