#!/bin/bash
cd /go/src/emotibot.com/emotigo/module/qic-api/;
go test ./... -json -cover > jenkins/test.json;

CONTENT=`cat jenkins/test.json | sed -e "$ ! s/$/,/g"`;
echo "const tests = [$CONTENT]" > jenkins/tests.js;