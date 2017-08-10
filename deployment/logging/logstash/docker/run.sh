#!/bin/bash
# NOTE:
# DO NOT touch anything outside <EDIT_ME></EDIT_ME>,
# unless you really know what you are doing.
REPO=docker-reg.emotibot.com.cn:55688
# The name of the container, should use the name of the repo is possible
# <EDIT_ME>
CONTAINER=logstash-jdbc
# </EDIT_ME>

# Get tags from args
TAG=$(git rev-list --abbrev-commit -1 HEAD -- "../../logstash")
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

ENV=$1
if [ "$ENV" != "dev" ] && [ "$ENV" != "idc" ] && [ "$ENV" != "vip" ] && [ "$ENV" != "changhong" ] && [ "$ENV" != "changhong.dev" ]; then
    echo "# Error: parameter does not satisfied."
    exit 1
fi

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BASE_DIR=$DIR/..

if [ ! -d "$BASE_DIR/backend_logs" ]; then
    mkdir -p "$BASE_DIR/backend_logs"
fi

JAR_DIR=$DIR/../plugin_jar

docker rm -f $CONTAINER
cmd="docker run --name $CONTAINER -td -v $BASE_DIR/backend_logs:/backend_logs -v $BASE_DIR/config-dir/logstash.$ENV.conf:/config-dir/logstash.conf -v $JAR_DIR:/vendor/jar -p 12201:12201/udp -p 5043:5043 $DOCKER_IMAGE -f /config-dir/logstash.conf"
echo $cmd
eval $cmd
