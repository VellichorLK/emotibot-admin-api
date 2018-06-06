#!/bin/bash

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
#default env file
envF="$DIR/default.env"
echo $#
if [ $# -eq 1 ]; then
    envF="$DIR/$1"
fi

cmd="docker run -it --env-file $envF --rm -v $DIR/data:/data docker-reg.emotibot.com.cn:55688/solr_verfiy:latest"
echo $cmd
eval $cmd