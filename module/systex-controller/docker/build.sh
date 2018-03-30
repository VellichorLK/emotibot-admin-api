#!/bin/bash
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
VERSION="latest"
if [ $# -eq 1 ];then
    VERSION=$1
fi
cmd="docker build -t docker-reg.emotibot.com.cn:55688/systex-controller:$VERSION -f $DIR/dockerfile $DIR/.."

echo $cmd
eval $cmd

