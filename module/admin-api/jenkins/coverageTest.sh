#!/bin/bash
cd /go/src/emotibot.com/emotigo/module/admin-api/;

gocov test ./... | gocov-xml > coverage.xml
