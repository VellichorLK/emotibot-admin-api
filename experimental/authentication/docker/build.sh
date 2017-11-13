#!/bin/bash
#REPO=docker-reg.emotibot.com.cn:55688
#CONTAINER=authentication
## TAG="$(git rev-parse --short HEAD)"
## TAG="20170926001"
#DATE=`date +%Y%m%d`
#TAG=$1
#
#LAST_RELEASE_TAG="20170926001"
#if [ "$TAG" == "" ]; then
#    TAG="$LAST_RELEASE_TAG"
#fi
#
#DOCKER_IMAGE=$REPO/$CONTAINER:$TAG
#
#DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
#EMOTIGO=${DIR#*/experimental/}
#BUILDROOT=${DIR%/experimental/*}
#GOSRCPATH="experimental/$EMOTIGO/../"
#
## Build docker
#cmd="docker build \
#  -t $DOCKER_IMAGE \
#  --build-arg PROJECT=$GOSRCPATH \
#  -f $DIR/Dockerfile $BUILDROOT"
#echo $cmd
#eval $cmd

REPO=docker-reg.emotibot.com.cn:55688
REPO_HARBOR=harbor.emotibot.com
PROJ_HARBOR=emotibot-k8s
CONTAINER=authentication

DATE=`date +%Y%m%d`
GIT_SHORT_HEAD="$(git rev-parse --short HEAD)"

TAG=$1
if [ "$TAG" == "" ]; then
    TAG="$DATE-$GIT_SHORT_HEAD"
fi

DOCKER_IMAGE=($REPO/$CONTAINER:$TAG $REPO_HARBOR/$PROJ_HARBOR/$CONTAINER:$TAG)

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
EMOTIGO=${DIR#*/experimental/}
BUILDROOT=${DIR%/experimental/*}
GOSRCPATH="experimental/$EMOTIGO/../"

for img in ${DOCKER_IMAGE[*]} 
do
    cmd="docker build -t $img \
        -f $DIR/Dockerfile \
        --build-arg PROJECT=$GOSRCPATH \
        $BUILDROOT"
    echo $cmd
    eval $cmd
done
