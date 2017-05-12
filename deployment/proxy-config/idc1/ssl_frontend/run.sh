#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

NAME='ssl_frontend-shadow'

docker rm -fv $NAME
docker run -d --name $NAME\
  --restart always \
  -p 9443:9443 \
  -v $DIR/haproxy.cfg:/usr/local/etc/haproxy/haproxy.cfg:ro \
  -v $DIR/cert/:/etc/ssl/ \
  haproxy:1.6

echo "# Starting $NAME on idc1"
docker ps | grep $NAME
