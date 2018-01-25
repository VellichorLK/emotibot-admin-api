#!/bin/bash

rounds=600
sleepTime=1

qps=$1
if [ -z "$qps" ]; then
    qps=1
fi

Server=$2
if [ -z "$Server" ]; then
  Server='172.16.101.122:15801'
fi

Userid=$3
if [ -z "$Userid" ]; then
  Userid="RANDOM"
fi

echo "Server: $Server"
echo "QPS: $qps"
echo "Userid: $Userid"

expressions=(
 "你们这群废柴给我收声！" \
 "口桀-口桀-朕還要再幹10個宮女" \
 "今次大獲仆街啦！" \
 "战吧！给我败啊！" \
 "口胡！今天我定要把你轰杀至渣啊！" \
 "你还未够班啊！" \
)

expressions1=(
 "国足12强赛赛程" \
 "国足12强赛赛程?" \
 "国足12强赛赛程2016" \
 "国足12强赛赛程3" \
 "国足12强赛赛程4" \
 "国足12强赛赛程5" \
 "国足12强赛赛程6" \
)

#curl -X POST -d\
#'cmd=chat&appid=facdbd02cc3324ccd9879b208a611e38&userid=0B19080F6B43D0B0CFC688C0D6225B793&\
#text=国足12强赛赛程$RANDOM'\
#http://idc.emotibot.com:80/api/ApiKey/openapi.php &

for (( i=1; i<=$rounds; i++ ))
do
  echo "Round $i"
  for (( j=1; j<=$qps; j++ )); do
    TXT=${expressions[$RANDOM % ${#expressions[@]} ]}
    #OpenAPI
    USERID=$Userid
    if [ "$USERID" = "RANDOM" ]; then
      USERID=$RANDOM
    fi
    echo "http://$Server/vip/irobot/ask4Json?userId=test2&platform=android&question=$TXT"
    time curl -v "http://$Server/vip/irobot/ask4Json?userId=test2&platform=android&question=$TXT"
  done
  sleep $sleepTime
done


