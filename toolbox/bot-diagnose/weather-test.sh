#!/bin/bash
rounds=1
sleepTime=1
qps=$1
if [ -z "$qps" ]; then
    qps=1
fi

Server=$2
if [ -z "$Server" ]; then
  Server='http://localhost:80'
fi

echo "Server: $Server"
echo "QPS: $qps"

expressions=(
 "台北天气" \
 )

for (( i=1; i<=$rounds; i++ ))
do
  echo "Round $i"
  for (( j=1; j<=$qps; j++ )); do
    TXT=${expressions[$RANDOM % ${#expressions[@]} ]}
    # solitaire
    # curl -H "Content-Type: application/json" -X POST -d '{"old": [], "input": "八拜之交"}' http://idc1:12280/api/idiom-solitaire &

    # Speechact
    # curl -H "Content-Type: application/json" -X POST -d '{"sentences":["信件已经寄出"]}' idc1:10280
    # Houta
    # echo 'http://idc1.emotibot.com:11180/api/APP/chat2.php?text=我要逆天'_$RANDOM'&wechatid=Test&type=text#' &
    # cmd="curl -v --get"
    cmd="curl -kv --get"
    cmd="$cmd --data-urlencode text=\"$TXT\""
    cmd="$cmd --data wechatid=Test"
    cmd="$cmd --data type=text"
    cmd="$cmd $Server/api/APP/chat2.php "
    echo $cmd
    eval time $cmd
  done
  sleep $sleepTime
done


# ./loadTesting.sh <qps> &> out.txt
# cat out.txt  | grep real | awk '{gsub("0m","",$2); gsub("s","",$2); print $2}'
