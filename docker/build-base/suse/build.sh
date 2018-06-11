#!/bin/bash
set -e

REPO=docker-reg.emotibot.com.cn:55688
CONTAINER=gobase-suse
TAG=1.9-alpine
DOCKER_IMAGE=$REPO/$CONTAINER:$TAG

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
printf $DOCKER_IMAGE > $DIR/DOCKER_IMAGE

cmd="docker build \
        -t $DOCKER_IMAGE \
        -f Dockerfile ."
echo $cmd
eval $cmd
