#!/bin/bash

#TEXT="我叫內湖金城武"
TEXT=$1
USERID=0B0D4C9C8AAB0407C9047CBB578B1CE9F
APPID=d1ed09ea950215aef0de3d649fcef1b4

curl -v -X POST \
 -d "cmd=chat&appid=$APPID&userid=$USERID&text=$TEXT"\
 http://idc.emotibot.com/api/ApiKey/openapi.php
