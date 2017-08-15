#!/bin/bash

SERVER_NAME=$1
if [ -z "$SERVER_NAME" ]; then
    echo "$0 content-sh.emotibot.com"
    exit 1
fi

if [ ! -d "$PWD/ssl.$SERVER_NAME" ]; then
    echo "$PWD/ssl.$SERVER_NAME does not exist!"
    exit 1 
fi

BACKEND_LOGS=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

container=nginx
tag=1.11.9

docker rm -f $container

cmd="docker run --name $container --log-opt max-size=20m --log-opt max-file=5 -d -v $PWD/ssl.$SERVER_NAME/Nginx/1_content-sh.emotibot.com_bundle.crt:/etc/nginx/ssl/nginx.crt  -v $PWD/ssl.$SERVER_NAME/Nginx/2_content-sh.emotibot.com.key:/etc/nginx/ssl/nginx.key -v $PWD/nginx.conf:/etc/nginx/nginx.conf -p 80:80 -p 443:443 $container:$tag"
echo $cmd
eval $cmd
