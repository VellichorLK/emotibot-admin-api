#!/bin/bash

export COMPOSE_PROJECT_NAME=qic-api-test

docker-compose -f integration.yaml up -d