#!/bin/bash
LELE_BACKUP_PORT=9011

ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Start the ssl termination proxy
echo "# Start ssl proxy"
cd $ROOT/ssl_frontend
./run.sh

# Start the golang uid header proxy
echo "# Start golang header proxy"
cd $ROOT/goproxy-uid
./run.sh

# Backup server when the incoming qps is too high
echo "# Start a backup server on port $BACKUP_SERV_PORT"
cd $ROOT/lesports-backup-nginx
./run.sh $LELE_BACKUP_PORT

# Start haproxy
echo "# Start haproxy on idc1"
cd $ROOT/haproxy-idc1
./run.sh


