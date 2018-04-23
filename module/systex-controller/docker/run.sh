#!/bin/bash

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
VERSION=`cat $DIR/VERSION`
PORT=$1
LOG_VOLUME_NAME=systex-log
VOLUME_EXIST=`docker volume ls -f name=$LOG_VOLUME_NAME -q | wc -l`
if [ $VOLUME_EXIST -eq 0 ]; then
    echo "Create a new volume: $LOG_VOLUME_NAME"
    docker volume create $LOG_VOLUME_NAME
fi
if [ $# -ge 2 ]; then
    VERSION=$2
    echo "Using version $VERSION"
fi
# docker pull docker-reg.emotibot.com.cn:55688/systex-controller:$VERSION
echo "remove old container if exists"
docker rm -f Systex-controller > /dev/null 2>&1

cmd="docker run -d --name Systex-controller \
    -e TZ=Asia/Taipei \
    --env-file $DIR/../.env \
    -p $PORT:80 \
    --log-opt max-size=20m \
    --log-opt max-file=4 \
    -v $LOG_VOLUME_NAME:/app/log/ \
    docker-reg.emotibot.com.cn:55688/systex-controller:$VERSION"
echo $cmd
eval $cmd