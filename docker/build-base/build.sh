#!/bin/bash
set -e
REPO=harbor.emotibot.com
PROJECT=library
CONTAINER=go-build
GOVER=1.10
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Read platform parameter
platforms=(`ls -d */ | sed -e "s/\///g"`)
if [ $# -ne 1 ]; then
  echo "Which platform you want to build ?"
  for i in "${!platforms[@]}"; do
    if [[ i -eq 0 ]]; then
      echo "$i: ${platforms[$i]} (default)";
    else
      echo "$i: ${platforms[$i]}";
    fi
  done
  read choice
  platform=${platforms[$choice]};
else
  platform=$1
fi

# check if platform setting is available
if [[ -z "$platform" ]]; then
  echo "Invalid platform";
  exit 1;
elif ! [[ -d "$platform" ]]; then
  echo "Invalid platform";
  exit 1
fi
echo "Build base for platform [$platform] with go ver [$GOVER]";

cd $platform && ./build.sh $GOVER;
if ! [[ $? -eq 0 ]]; then
  echo "Build base image of platform fail";
  exit 1;
fi

if [ ! -f "$DIR/$platform/DOCKER_IMAGE" ]; then
  echo "$DIR/$platform/DOCKER_IMAGE does not exist!"
  exit 1
fi

baseImageName=`cat $DIR/$platform/DOCKER_IMAGE`
echo "building base on $baseImageName"


if [ "$(uname)" == "Darwin" ]; then
  # Use mac command
  VERSION=`shasum $DIR/Dockerfile | awk '{ print $1 }'| cut -c1-8`
elif [ "$(expr substr $(uname -s) 1 5)" == "Linux" ]; then
  # Use linux command
  VERSION=`(sha1sum $DIR/Dockerfile) | awk '{ print $1 }'| cut -c1-8`
else
  echo "unsupport platform"
  exit
fi
TAG="$VERSION-$GOVER-$platform"
DOCKER_IMAGE=$REPO/$PROJECT/$CONTAINER:$TAG

BUILDROOT=$DIR
# Build docker
cmd="docker build \
  -t $DOCKER_IMAGE \
  --build-arg baseImageName=$baseImageName
  -f $DIR/Dockerfile $BUILDROOT"
echo $cmd
eval $cmd
if ! [[ -z $? ]]; then
  echo "Build docker $DOCKER_IMAGE success"
else
  echo "Build docker $DOCKER_IMAGE fail"
fi
