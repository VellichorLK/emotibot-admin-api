#!/bin/bash
set -e

export COMPOSE_PROJECT_NAME=qic-api-test
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
docker-compose -p $COMPOSE_PROJECT_NAME -f $DIR/integration.yaml up -d -V --force-recreate
containerID=$(docker ps -qf "name=mysql-integration")
until [ "$(docker inspect -f {{.State.Health.Status}} mysql-integration)" == "healthy" ]; do
    echo "wait-for-it";
    sleep 3;
done;
docker exec -i mysql-integration mysql -u root -ppassword -h 127.0.0.1 < $DIR/data.sql
docker cp $DIR/call.csv $containerID:/call.csv
docker exec -it mysql-integration mysqlimport --local --ignore-lines=1 --fields-terminated-by=, --fields-optionally-enclosed-by=\" -h 127.0.0.1 --user root -ppassword QISYS /call.csv
docker cp $DIR/task.csv $containerID:/task.csv
docker exec -it mysql-integration mysqlimport --local --ignore-lines=1 --fields-terminated-by=, --fields-optionally-enclosed-by=\" -h 127.0.0.1 --user root -ppassword QISYS /task.csv
cd ../ && go test "$@" ./... -integration

docker-compose -p $COMPOSE_PROJECT_NAME -f $DIR/integration.yaml down -t 3