#!/bin/bash
set -e

export COMPOSE_PROJECT_NAME=qic-api-test
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
docker-compose -p $COMPOSE_PROJECT_NAME -f $DIR/integration.yaml up -d -V --force-recreate
until [ "$(docker inspect -f {{.State.Health.Status}} mysql-integration)" == "healthy" ]; do
    echo "wait-for-it";
    sleep 3;
done;
mysql -u root -ppassword -h 127.0.0.1 < $DIR/data.sql
mysqlimport --local --ignore-lines=1 --fields-terminated-by=, -h 127.0.0.1 --user root -ppassword QISYS ./call.csv
cd ../ && go test "$@" ./... -integration

docker-compose -p $COMPOSE_PROJECT_NAME -f $DIR/integration.yaml down -t 3