#!/bin/bash

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

VERSION=`cat $DIR/VERSION`
$DIR/build.sh $VERSION
cmd="docker push docker-reg.emotibot.com.cn:55688/systex-controller:$VERSION"
echo $cmd
eval $cmd

