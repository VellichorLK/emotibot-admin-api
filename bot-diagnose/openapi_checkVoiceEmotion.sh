#!/bin/bash
set -x
HOST=idc.emotibot.com:80
#HOST=$1
CMD=checkVoiceEmotion
APPID=5a200ce8e6ec3a6506030e54ac3b970e
USERID=1
DIR=$(pwd)
curl -v \
    -X POST \
    -F "cmd=$CMD" \
    -F "appid=$APPID" \
    -F "userid=$USERID" \
    -F "file=@$DIR/test.amr" \
    -F "type=man" \
    http://$HOST/api/ApiKey/openapi.php


#curl -v \
#http://$HOST/Files/voice/2043119.mp3?userid=$USERID&appid=5a200ce8e6ec3a6506030e54ac3b970e
