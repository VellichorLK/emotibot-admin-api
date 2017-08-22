#!/bin/bash

REPO=docker-reg.emotibot.com.cn:55688
# The name of the container, should use the name of the repo is possible
CONTAINER=authentication-test
TAG="20170821001"
DOCKER_IMAGE=$REPO/$CONTAINER:$TAG

cmd="docker run --rm --name $CONTAINER $DOCKER_IMAGE newman run -e $1 auth_server_enterprise.postman_collection.json"

echo $cmd
eval $cmd
