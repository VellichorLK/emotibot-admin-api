#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd $DIR/haproxy-idc1
pwd && ./kill.sh
cd $DIR/goproxy-uid
pwd && ./kill.sh
cd $DIR/lesports-backup-nginx
pwd && ./kill.sh
cd $DIR/ssl_frontend
pwd && ./kill.sh
