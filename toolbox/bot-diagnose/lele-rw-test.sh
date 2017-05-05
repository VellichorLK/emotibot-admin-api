#!/bin/bash

HOST=sta1.emotibot.com.cn
HOST=bot.emotibot.com

# The user id in sta1's db
USERID=035BA5AC8E856531595F0FC84D2AF3BBA

# A random userID
USERID=5566

cmd="curl -X POST -d \"cmd=chat&appid=facdbd02cc3324ccd9879b208a611e38&userid=$USERID&text=张呈栋多大了\" http://$HOST/api/ApiKey/openapi.php"
echo $cmd
eval $cmd
# 27

echo ""

cmd="curl -X POST -d \"cmd=chat&appid=facdbd02cc3324ccd9879b208a611e38&userid=$USERID&text=他中超进了几球\" http://$HOST/api/ApiKey/openapi.php"
echo $cmd
eval $cmd


