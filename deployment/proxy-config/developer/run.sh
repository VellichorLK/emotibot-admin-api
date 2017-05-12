#!/bin/bash
# REPO=docker-reg.emotibot.com.cn:55688
CONTAINER=ssl-frontend
TAG=$2
DOCKER_IMAGE=haproxy:1.6

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Load the env file
source $1
if [ $? -ne 0 ]; then
  if [ "$#" -eq 0 ];then
    echo "Usage: $0 <envfile>"
    echo "e.g., $0 dev.env"
  else
    echo "Erorr, can't open envfile: $1"
  fi
  exit 1
else
  echo "# Using envfile: $1"
fi

# Start docker
docker rm -f -v $CONTAINER
cmd="docker run -d --name $CONTAINER \
  --restart="always" \
  -p 443:443 \
  -v $DIR/cert/$HAPROXY_CERT:/etc/ssl/emotibot_com.pem:ro \
  -v $DIR/conf/$HAPROXY_CFG:/usr/local/etc/haproxy/haproxy.cfg:ro \
  $DOCKER_IMAGE \
"

echo $cmd
eval $cmd
