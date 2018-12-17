#!/bin/bash

export COMPOSE_PROJECT_NAME=emotion-engine-test

docker-compose -f ./emotion-engine-test.yaml up -d --force-recreate