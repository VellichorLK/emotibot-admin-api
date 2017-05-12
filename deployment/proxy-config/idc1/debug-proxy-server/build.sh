#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "# Building all required images first"
cd $DIR/../goproxy-uid
./build.sh

echo "# Pulling haproxy docker"
docker pull haproxy:1.6
