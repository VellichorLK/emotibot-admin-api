#!/bin/bash
# NOTE: parts of the script is automatically generated.
# DO NOT edit the build header and build footer parts.
## <Build Header> ##
REPO=docker-reg.emotibot.com.cn:55688
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

BUILDROOT=$DIR/../../
## </Build Header> ##
### Begin of your customized build scripts ###

CONTAINER=python-template-worker

### End of your customized build scripts ###
## <Build Footer> ##

REPO=docker-reg.emotibot.com.cn:55688
TAG=20170518
DOCKER_IMAGE=$REPO/$CONTAINER:$TAG

cmd="docker build \
  -t $DOCKER_IMAGE \
  -f $DIR/Dockerfile \
  $BUILDROOT"
echo $cmd
eval $cmd
## </Build Footer> ##
