#!/bin/bash
REPO=docker-reg.emotibot.com.cn:55688
CONTAINER=logstash-jdbc
TAG=$(git rev-list --abbrev-commit -1 HEAD -- "../../logstash")
DOCKER_IMAGE=$REPO/$CONTAINER:$TAG

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BUILDROOT=$DIR/..


# Build docker
cmd="docker build -t $DOCKER_IMAGE -f $DIR/Dockerfile $BUILDROOT"
echo $cmd
eval $cmd
