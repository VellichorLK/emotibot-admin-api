#!/bin/bash

pushd rabbitmq/docker && ./run.sh test.env
popd


types="python java golang"
replicas=2

for language in $types; do

	env_file="workers/${language}_worker/docker/test.env"
	container="${language}-template-worker"
	tag=20170518
	img="docker-reg.emotibot.com.cn:55688/$container:$tag"
	echo $img
	for (( i=0; i<$replicas; i=i+1 ))
	do
		cmd="docker run -d --name $container${i}   -v /etc/localtime:/etc/localtime --restart always --env-file $env_file $img"
		echo $cmd
		eval $cmd
	done
done


pushd web_service/docker && ./run.sh test.env
popd