#!/bin/bash

BACKEND_LOGS=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

if [ ! -f "$BACKEND_LOGS/kibana_htpasswd" ]; then
    exit 1;
fi

container=nginx
tag=1.11.9

docker rm -f $container
cmd="docker run --name $container -d -v $PWD/nginx.conf:/etc/nginx/nginx.conf -v $PWD/kibana_htpasswd:/etc/nginx/kibana_htpasswd -p 8088:80 $container:$tag"
echo $cmd
eval $cmd
