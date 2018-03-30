#!/bin/bash

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
VERSION=`cat $DIR/VERSION`
PORT=$1
# docker pull docker-reg.emotibot.com.cn:55688/systex-controller:$VERSION
docker rm -f Systex-controller

docker run -d --name Systex-controller \
    --env-file $DIR/../.env \
    -p $PORT:80 \
    --log-opt max-size=20m \
    --log-opt max-file=4 \
    docker-reg.emotibot.com.cn:55688/systex-controller:$VERSION