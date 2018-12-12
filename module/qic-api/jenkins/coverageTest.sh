#!/bin/bash
cd /go/src/emotibot.com/emotigo/module/qic-api/;

gocov test ./... | gocov-xml > coverage.xml
