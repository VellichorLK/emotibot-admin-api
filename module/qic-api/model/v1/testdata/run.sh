#!/bin/bash
set -e

export COMPOSE_PROJECT_NAME=qic-api-test
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
docker-compose -p $COMPOSE_PROJECT_NAME -f $DIR/integration.yaml up -d -V --force-recreate sql
cID=$(docker-compose -p $COMPOSE_PROJECT_NAME -f $DIR/integration.yaml ps -q sql)
docker-compose -p $COMPOSE_PROJECT_NAME -f $DIR/integration.yaml run init-db entrypoint.sh

importfiles=$(find $(pwd) -regex ".*\.\(csv\)")
for f in $importfiles;
do
    basename=${f##*/}
    docker cp $f $cID:/$basename
    docker exec -it mysql-integration mysqlimport --local --ignore-lines=1 --fields-terminated-by=, --fields-optionally-enclosed-by=\" -h 127.0.0.1 --user root -ppassword QISYS /$basename
done

docker-compose -p $COMPOSE_PROJECT_NAME -f $DIR/integration.yaml down -t 3