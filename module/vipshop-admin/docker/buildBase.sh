#!/bin/bash

GIT_HEAD="$(git rev-parse --short HEAD)"
IMAGE="docker-reg.emotibot.com.cn:55688/vipshop-admin-api:$GIT_HEAD-benchmarkOnly"
BUILDROOT=$PWD/../../
MODULE="vipshop-admin"
echo "Module:$MODULE"
docker build -t $IMAGE -f baseDockerfile   --build-arg PROJECT=$MODULE $BUILDROOT