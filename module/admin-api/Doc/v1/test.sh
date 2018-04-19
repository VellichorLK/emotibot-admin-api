#!/bin/bash
folder=${PWD}

docker run \
    -v $folder:/spec \
    -v $folder/target:/gen \
    -e "LANGUAGE=html" \
    -e "SWAGGER_FILE=qa.yaml" \
    sandcastle/swagger-codegen-docker
