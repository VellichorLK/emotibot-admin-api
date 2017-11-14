#!/bin/bash
REPO=docker-reg.emotibot.com.cn:55688
CONTAINER=vipshop-admin-api
# TAG="$(git rev-parse --short HEAD)"
TAG="2017111404"
DOCKER_IMAGE=$REPO/$CONTAINER:$TAG

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
EMOTIGO=${DIR#*/module/}
BUILDROOT=${DIR%/module/*}
GOSRCPATH="module/$EMOTIGO/../"

# Build docker
cmd="docker build \
  -t $DOCKER_IMAGE \
  --build-arg PROJECT=$GOSRCPATH \
  -f $DIR/Dockerfile $BUILDROOT"
echo $cmd
eval $cmd
