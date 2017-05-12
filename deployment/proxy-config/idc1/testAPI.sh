#!/bin/bash

echo "hi"
HOST=$1
if [ -z "$HOST" ]; then
    HOST='127.0.0.1'
fi
PORT=$2
if [ -z "$PORT" ]; then
    PORT=9010
fi

URL=api/ApiKey/openapi.php

echo "# Test lele cluster"
time curl -v -X POST -d \
"cmd=chat&appid=facdbd02cc3324ccd9879b208a611e38&userid=0B19080F6B43D0B0CFC688C0D6225B793&text=国足12强赛赛程$RANDOM" \
http://$HOST:$PORT/$URL
