#!/bin/bash

PORT=$1
if [ -z "$PORT" ]; then
  echo "Usage: $0 <port>"
  exit 1
fi

NAME=lesports-backup-nginx-$PORT

docker rm -fv $NAME
docker run --name $NAME \
  -v $(pwd)/html:/usr/share/nginx/html \
  -v $(pwd)/default.conf:/etc/nginx/conf.d/default.conf \
  -p $PORT:80 -d nginx:alpine 

echo "curl localhost:$PORT"
echo "docker rm -fv $NAME"
