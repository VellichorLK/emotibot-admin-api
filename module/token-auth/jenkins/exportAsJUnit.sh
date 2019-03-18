#!/bin/bash
cd /go/src/emotibot.com/emotigo/module/token-auth/;

go test -v ./... 2>&1 | go-junit-report > jenkins/unittest.xml
