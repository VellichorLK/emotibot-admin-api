#!/bin/bash
REPO=docker-reg.emotibot.com.cn:55688
CONTAINER=vipshop-admin-api
# TAG="$(git rev-parse --short HEAD)"
LAST_RELEASE_TAG="20171124-fe27a1f"
GIT_HEAD="$(git rev-parse --short HEAD)"
DATE=`date +%Y%m%d`
TAG=$1

if [ "$TAG" == "" ]; then
    TAG="$DATE-$GIT_HEAD"
elif [ "$TAG" == "LR" ]; then
    TAG=$LAST_RELEASE_TAG
fi

DOCKER_IMAGE=$REPO/$CONTAINER:$TAG

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
GOSRCPATH="$(cd "$DIR/../" && pwd )"
MODULE=${GOSRCPATH##/*/}
BUILDROOT=$DIR/../../

# Build docker
cmd="docker build \
  -t $DOCKER_IMAGE \
  --build-arg PROJECT=$MODULE \
  -f $DIR/Dockerfile $BUILDROOT"
echo $cmd
eval $cmd
