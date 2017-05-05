#!/bin/bash

function push_docker(){
  cmd="docker tag $1 $2"
  cmd2="docker push $2"
  cmd3="docker rmi $2"
  echo $cmd; eval $cmd
  echo $cmd2; eval $cmd2
  sleep 3
  echo $cmd3; eval $cmd3
}

# Check if the args are good
if [ $# -eq 0 ]
  then
    echo "# Usage: $0 <docker image>"
    echo "# e.g., "
    echo "#   $0 docker-reg.emotibot.com.cn:55688/speechact:94crazy"
    echo "#   Will pull the image from SH and then push to idc's registry"
    exit 1
fi

# Check if we are running the script in the right host
# 1. must in IDC
# 2. docker-reg.emotibot.com.cn set to SH office's ip 180.x
REGIP=$(ping -c 1 docker-reg.emotibot.com.cn | grep PING | grep 180)
if [ "$REGIP" == "" ]; then
    echo "# Error, you need to run the script on specialized hosts, e.g., idc23"
    exit 1
fi


ONE_IMG=$1
REPO_IDC=idc-docker.emotibot.com.cn:55688

if [ "$service" == "all" ]; then
	IMG_LIST=$(docker ps --format={{.Image}} | grep docker-reg)
	for i in $IMG_LIST
	do
		CNAME=`echo $i | cut -d '/' -f2`
		new_tag=$REPO_IDC/$CNAME
		push_docker $i $new_tag
	done
else
    IMG=$ONE_IMG
    docker pull $IMG
    CNAME=`echo $IMG | cut -d '/' -f2`
    new_tag=$REPO_IDC/$CNAME
    push_docker $IMG $new_tag
fi
