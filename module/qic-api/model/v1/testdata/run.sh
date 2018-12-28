#!/bin/bash

export COMPOSE_PROJECT_NAME=qic-api-test
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
docker-compose -p $COMPOSE_PROJECT_NAME -f $DIR/integration.yaml up -d
until [ "$(docker inspect -f {{.State.Health.Status}} mysql-integration)" == "healthy" ]; do
    echo "wait-for-it";
    sleep 1;
done;
mysql -u root -ppassword -h 127.0.0.1 < $DIR/data.sql
cd ../ && go test -v ./...-integration

docker-compose -p $COMPOSE_PROJECT_NAME -f $DIR/integration.yaml down -t 3