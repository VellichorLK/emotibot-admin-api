#!/bin/bash
cd /go/src/emotibot.com/emotigo/module/token-auth/;

gocov test ./... | gocov-xml > coverage.xml
