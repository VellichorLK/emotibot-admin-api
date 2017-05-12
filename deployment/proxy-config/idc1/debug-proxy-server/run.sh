#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

NAME='haproxy-idc1-debug'

docker rm -fv $NAME
docker run -d --name $NAME\
  --restart always \
  -p 9001:9001 \
  -v $DIR/haproxy.demo.cfg:/usr/local/etc/haproxy/haproxy.cfg:ro \
  haproxy:1.6

echo "# Starting a demo haproxy on idc1"
sleep 1
docker ps | grep haproxy-idc


echo "# Starting all demo goservers"
Servers="46 47 48 50 51 52 53 54 55 56"

for s in $Servers; do
  # echo $s idc$s 100$s
  name="idc$s"
  port=100$s
  docker rm -fv $name
  docker run -d --name $name -p $port:$port \
    docker-reg.emotibot.com.cn:55688/goproxy-uid:20170512 \
    /app/fakeserver/fakeserver $port $name
done

echo "# Start the goproxy"
cd $DIR/../goproxy-uid
./run.sh

echo "# Your debug proxy should be good now."
curl -v localhost:9000?userid=5566
