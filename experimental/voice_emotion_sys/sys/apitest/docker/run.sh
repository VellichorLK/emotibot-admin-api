#!/bin/bash
# NOTE:
# DO NOT touch anything outside <EDIT_ME></EDIT_ME>,
# unless you really know what you are doing.

REPO=docker-reg.emotibot.com.cn:55688
# The name of the container, should use the name of the repo is possible
# <EDIT_ME>
CONTAINER=tester
# </EDIT_ME>

# Get tags from args
TAG=20170801
DOCKER_IMAGE=$REPO/$CONTAINER:$TAG
echo "# Launching $DOCKER_IMAGE"
# Check if docker image exists (locally or on the registry)
local_img=$(docker images | grep $REPO | grep $CONTAINER | grep $TAG)
if [ -z "$local_img" ] ; then
  echo "# Image not found locally, let's try to pull it from the registry."
  docker pull $DOCKER_IMAGE
  if [ "$?" -ne 0 ]; then
    echo "# Error: Image not found: $DOCKER_IMAGE"
    exit 1
  fi
fi
echo "# Great! Docker image found: $DOCKER_IMAGE"



# global config:
# - use local timezone
# - max memory = 5G
# - restart = always
globalConf="
  -v /etc/localtime:/etc/localtime \
  --restart always \
  --log-opt max-size=20m \
  --log-opt max-file=20 \
"

# <EDIT_ME>
moduleConf="
  --env-file $envfile \
"
# </EDIT_ME>

docker rm -f -v $CONTAINER
cmd="docker run -t --name $CONTAINER \
  $globalConf \
  $moduleConf \
  $DOCKER_IMAGE \
  $@ \
"
echo $cmd
exit 0
eval $cmd
