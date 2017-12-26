#!/bin/bash
REPO=docker-reg.emotibot.com.cn:55688
CONTAINER=go-build
DATE=`date +%Y%m%d`
GITHEAD="$(git rev-parse --short HEAD)"
TAG=${DATE}_${GITHEAD}

DOCKER_IMAGE=$REPO/$CONTAINER:$TAG

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
GOSRCPATH="$(cd "$DIR/../" && pwd )"
MODULE=${GOSRCPATH##/*/}
BUILDROOT=$DIR/../../

# Build docker
cmd="docker build \
  -t $DOCKER_IMAGE \
  -f $DIR/Dockerfile $BUILDROOT"
echo $cmd
eval $cmd
