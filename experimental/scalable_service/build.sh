#!/bin/bash


pushd rabbitmq/docker && ./build.sh
popd && pushd web_service/docker && ./build.sh
popd && pushd workers/python_worker/docker && ./build.sh
popd && pushd workers/golang_worker/docker && ./build.sh
popd && pushd workers/java_worker/docker && ./build.sh