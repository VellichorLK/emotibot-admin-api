#!/bin/bash

#!/bin/bash
REPO=docker-reg.emotibot.com.cn:55688
# The name of the container, should use the name of the repo is possible
CONTAINER=vipshop-fake-cas
TAG="1.0.2"
DOCKER_IMAGE=$REPO/$CONTAINER:$TAG

globalConf="
    -v /etc/localtime:/etc/localtime \
    --restart always \
"
moduleConf="
    -p 8788:443
"

docker rm -f $CONTAINER
cmd="docker run -d --name $CONTAINER \
    $globalConf \
    $moduleConf \
    $DOCKER_IMAGE \
"

echo $cmd
eval $cmd
