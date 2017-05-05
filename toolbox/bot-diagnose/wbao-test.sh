#!/bin/bash

rounds=600
sleepTime=1
# wbao appid
Appid="c385e97b0cdce3bdbbee59083ec3b0d0"

qps=$1
if [ -z "$qps" ]; then
    qps=1
fi

Server=$2
if [ -z "$Server" ]; then
  Server='wangbao.emotibot.com:80'
fi

Userid=$3
if [ -z "$Userid" ]; then
  Userid="RANDOM"
fi

echo "Server: $Server"
echo "QPS: $qps"
echo "Userid: $Userid"
echo "Appid: $Appid"
expressions=(
"电影真好看" \
"我好兴奋啊" \
"我失恋了" \
)
expressions1=(
 "你们这群废柴给我收声！" \
 "口桀-口桀-朕還要再幹10個宮女" \
 "今次大獲仆街啦！" \
 "战吧！给我败啊！" \
 "口胡！今天我定要把你轰杀至渣啊！" \
 "你还未够班啊！" \
)

for (( i=1; i<=$rounds; i++ ))
do
  echo "Round $i"
  for (( j=1; j<=$qps; j++ )); do
    TXT=${expressions[$RANDOM % ${#expressions[@]} ]}
    #OpenAPI
    APPID=$Appid
    USERID=$Userid
    if [ "$USERID" = "RANDOM" ]; then
      USERID=$RANDOM
    fi
    TXTALL="cmd=chat&appid=$APPID&userid=$USERID&text=$TXT"
    echo $TXTALL
    cmd="curl -X POST -d \"$TXTALL\" http://$Server/api/ApiKey/openapi.php"
    echo $cmd
    eval "time $cmd "
  done
  sleep $sleepTime
done


