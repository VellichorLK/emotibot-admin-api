#!/bin/bash

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
VERSION=`cat $DIR/VERSION`
PORT=$1
LOG_VOLUME_NAME=systex-log
VOLUME_EXIST=`docker volume ls -f name=$LOG_VOLUME_NAME -q | wc -l`
if [ $VOLUME_EXIST -eq 0 ]; then
    docker volume create $LOG_VOLUME_NAME
fi
docker pull docker-reg.emotibot.com.cn:55688/systex-controller:$VERSION
docker rm -f Systex-controller 2> /dev/null

docker run -d --name Systex-controller \
    -e TZ=Asia/Taipei \
    --env-file $DIR/../.env \
    -p $PORT:80 \
    --log-opt max-size=20m \
    --log-opt max-file=4 \
    -v $LOG_VOLUME_NAME:/app/log/ \
    docker-reg.emotibot.com.cn:55688/systex-controller:$VERSION