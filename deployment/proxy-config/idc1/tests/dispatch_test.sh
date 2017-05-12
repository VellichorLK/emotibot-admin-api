#!/bin/bash

HOST=$1
# POST
# curl -X POST --data userid=userid --data UserID=UserID2 http://localhost:9080

# GET
#curl 'http://localhost:9000/app/API/chat.php?xxx=123&UserID=UserID&TEXT=123&userid=userid'

# Test Image entries:
# curl -X POST --data userid=userid --data UserID=UserID2 http://$HOST/api/APP/getFashionxxx.php
# curl "http://$HOST/api/APP/getFashionxxx.php?data1=x&userid=xxx"

# Test openapi header
# curl -X POST --data userid=uid --data cmd=mycmd "http://$HOST/foo/bar"
# curl -F "userid=uid" -F "cmd=mycmd" "http://$HOST/foo/bar"
# curl "http://$HOST/foo/bar?userid=uid&appid=app1&cmd=getFace"
