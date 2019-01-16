#!/bin/bash
set -e

export COMPOSE_PROJECT_NAME=qic-api-test
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
COMPOSE_FILE_PATH=$DIR/docker-compose.yml

docker-compose -p $COMPOSE_PROJECT_NAME -f $COMPOSE_FILE_PATH  up -d -V --force-recreate
until [ "$(docker inspect -f {{.State.Health.Status}} rabbit)" == "healthy" ]; do
    echo "wait-for-it";
    sleep 3;
done;
cd $DIR/.. && go test "$@" ./... -integration
docker-compose -p $COMPOSE_PROJECT_NAME -f $COMPOSE_FILE_PATH down -t 3