#!/bin/bash
# NOTE:
# DO NOT touch anything outside <EDIT_ME></EDIT_ME>,
# unless you really know what you are doing.

REPO=docker-reg.emotibot.com.cn:55688
# The name of the container, should use the name of the repo is possible
# <EDIT_ME>
CONTAINER=goproxy-uid
# </EDIT_ME>
TAG="$(git rev-parse --short HEAD)"
DOCKER_IMAGE=$REPO/$CONTAINER:$TAG

# global config:
# - use local timezone
# - restart = always
globalConf="
  -v /etc/localtime:/etc/localtime \
  --restart always \
"
# <EDIT_ME>
moduleConf="
  -p 9000:9000\
"
# </EDIT_ME>

docker rm -f -v $CONTAINER
cmd="docker run -d --name $CONTAINER \
  $globalConf \
  $moduleConf \
  $DOCKER_IMAGE \
"
echo $cmd
eval $cmd
