#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

NAME='haproxy-idc1'

docker rm -fv $NAME
cmd="docker run -d --name $NAME\
  --restart always \
  -p 9001:9001 \
  -p 9010:9010 \
  -p 9527:9527 \
  -v $DIR/haproxy.default.cfg:/usr/local/etc/haproxy/haproxy.cfg:ro \
  haproxy:1.6"
 
echo $cmd
exec $cmd

echo "# Starting haproxy on idc1"
head -n 12 haproxy.cfg
docker ps | grep haproxy-idc
