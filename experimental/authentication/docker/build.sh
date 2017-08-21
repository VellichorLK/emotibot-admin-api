#!/bin/bash
REPO=docker-reg.emotibot.com.cn:55688
CONTAINER=authentication
# TAG="$(git rev-parse --short HEAD)"
TAG="20170821001"
DOCKER_IMAGE=$REPO/$CONTAINER:$TAG

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
EMOTIGO=${DIR#*/experimental/}
BUILDROOT=${DIR%/experimental/*}
GOSRCPATH="experimental/$EMOTIGO/../"

# Build docker
cmd="docker build \
  -t $DOCKER_IMAGE \
  --build-arg PROJECT=$GOSRCPATH \
  -f $DIR/Dockerfile $BUILDROOT"
echo $cmd
eval $cmd
