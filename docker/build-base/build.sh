#!/bin/bash
set -e
REPO=docker-reg.emotibot.com.cn:55688
CONTAINER=go-build
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [ $# -ne 1 ]; then
  echo "usage: ./build.sh alpine"
  exit 1
fi

if [ ! -f "$DIR/$1/DOCKER_IMAGE" ]; then
  echo "$DIR/$1/DOCKER_IMAGE does not exist!"
  exit 1
fi

baseImageName=`cat $DIR/$1/DOCKER_IMAGE`
echo "building base on $baseImageName"
if [ "$(uname)" == "Darwin" ]; then
  VERSION=`shasum $DIR/Dockerfile | awk '{ print $1 }'| cut -c1-8`
elif [ "$(expr substr $(uname -s) 1 5)" == "Linux" ]; then
  VERSION=$(sha1sum $DIR/Dockerfile)
else
  echo "unsupport platform"
  exit
fi
TAG="$VERSION-$1"
printf $VERSION > $DIR/VERSION
DOCKER_IMAGE=$REPO/$CONTAINER:$TAG

BUILDROOT=$DIR/..
# Build docker
cmd="docker build \
  -t $DOCKER_IMAGE \
  --build-arg baseImageName=$baseImageName
  -f $DIR/Dockerfile $BUILDROOT"
echo $cmd
eval $cmd