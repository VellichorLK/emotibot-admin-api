#!/bin/bash
# NOTE:
# DO NOT touch anything outside <EDIT_ME></EDIT_ME>,
# unless you really know what you are doing.

REPO=docker-reg.emotibot.com.cn:55688
# The name of the container, should use the name of the repo is possible
# <EDIT_ME>
CONTAINER=transfer
# </EDIT_ME>

# Get tags from args
TAG=2017102000
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

# Load env file
source $1
if [ "$?" -ne 0 ]; then
  echo "Error, can't open envfile: $1"
  echo "Usage: $0 <env file> <docker image tag>"
  echo "e.g., "
  echo " $0 dev.env 94crazy"
  exit 1
else
  envfile=$1
  echo "# Using envfile: $envfile"
fi

# global config:
# - use local timezone
# - max memory = 5G
# - restart = always
globalConf="
  -v /etc/localtime:/etc/localtime \
  --log-opt max-size=20m \
  --log-opt max-file=20 \
"

# <EDIT_ME>
moduleConf="
  --env-file $envfile \
"
# </EDIT_ME>

params="-i=$source -ipo=$source_port -ip=$source_pass -iu=$source_user -o=$destination -opo=$destination_port -ou=$destination_user -op=$destination_pass"

docker rm -f -v $CONTAINER
cmd="docker run -it --name $CONTAINER \
  $globalConf \
  $moduleConf \
  $DOCKER_IMAGE $params\
"
echo $cmd
eval $cmd
