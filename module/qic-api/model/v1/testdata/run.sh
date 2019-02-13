#!/bin/bash
set -e

export COMPOSE_PROJECT_NAME=qic-api-test
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
docker-compose -p $COMPOSE_PROJECT_NAME -f $DIR/integration.yaml up -d -V --force-recreate sql
cID=$(docker-compose -p $COMPOSE_PROJECT_NAME -f $DIR/integration.yaml ps -q sql)
docker-compose -p $COMPOSE_PROJECT_NAME -f $DIR/integration.yaml run init-db entrypoint.sh

csvfiles=$(find $(pwd) -regex ".*\.\(csv\)")
for f in $csvfiles;
do
    basename=${f##*/}
    docker cp $f $cID:/$basename
    docker exec -it mysql-integration mysqlimport --local --ignore-lines=1 --fields-terminated-by=, --fields-optionally-enclosed-by=\" -h 127.0.0.1 --user root -ppassword QISYS /$basename
done
tsvfiles=$(find $(pwd) -regex ".*\.\(tsv\)")
for f in $tsvfiles;
do
    basename=${f##*/}
    docker cp $f $cID:/$basename
    docker exec -it mysql-integration mysqlimport --local --ignore-lines=1 --fields-terminated-by=\\t --fields-optionally-enclosed-by=\" -h 127.0.0.1 --user root -ppassword QISYS /$basename
done
cd ../ && go test "$@" ./... -integration
docker-compose -p $COMPOSE_PROJECT_NAME -f $DIR/integration.yaml down -t -v --remove-orphans 3 > /dev/null 2>&1