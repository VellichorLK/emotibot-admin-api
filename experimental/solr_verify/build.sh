#!/bin/bash
set -ev
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

docker build -t docker-reg.emotibot.com.cn:55688/solr_verfiy --build-arg module_type=experimental --build-arg module_name=solr_verify -f dockerfile $DIR/../../