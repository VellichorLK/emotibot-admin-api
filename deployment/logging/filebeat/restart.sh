#!/bin/bash
REPO=docker-reg.emotibot.com.cn:55688
CONTAINER=filebeat
TAG=5.1.2
HOSTNAME=$( hostname )
BACKEND_LOGS_DIR="/home/deployer/backend_logs"
DOCKER_IMG=$REPO/$CONTAINER:$TAG

ENV=$1
if [ "$ENV" != "dev" ] && [ "$ENV" != "idc" ] && [ "$ENV" != "localhost" ] && [ "$ENV" != "vip" ]; then
    echo "invalid parameter"
    exit 1
fi

docker rm -f $CONTAINER
cmd="docker run --name $CONTAINER -d -v $PWD/filebeat.$ENV.yml:/filebeat.yml -v $BACKEND_LOGS_DIR:/var/logs/filebeat $DOCKER_IMG -E name=$HOSTNAME"
echo $cmd
eval $cmd
