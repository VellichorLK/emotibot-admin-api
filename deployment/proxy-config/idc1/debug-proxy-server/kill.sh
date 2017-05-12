#!/bin/bash

echo "# Stop haproxy"
NAME="goproxy-uid"
docker rm -fv $(docker ps -q --filter "name=$NAME")

echo "# Kill all fakeserver(s)"
NAME="idc"
docker rm -fv $(docker ps -q --filter "name=$NAME")

