#!/bin/sh

GIT_HEAD="$(git rev-parse --short=7 HEAD)"
GIT_DATE=$(git log HEAD -n1 --pretty='format:%cd' --date=format:'%Y%m%d-%H%M')

# Use docker-reg for now, it will change to harbor after harbor usage is confirmed
# OPENAPI_ADAPTER_REPO="harbor.emotibot.com/bfop"
OPENAPI_ADAPTER_REPO="docker-reg.emotibot.com.cn:55688"
OPENAPI_ADAPTER_CONTAINER="openapi-adapter"
OPENAPI_ADAPTER_TAG="${GIT_HEAD}_${GIT_DATE}"

OPENAPI_ADAPTER_BUILD_IMAGE_NAME="${OPENAPI_ADAPTER_REPO}/${OPENAPI_ADAPTER_CONTAINER}-build:${OPENAPI_ADAPTER_TAG}"
OPENAPI_ADAPTER_IMAGE_NAME="${OPENAPI_ADAPTER_REPO}/${OPENAPI_ADAPTER_CONTAINER}:${OPENAPI_ADAPTER_TAG}"
OPENAPI_ADAPTER_CONTAINER_NAME=${OPENAPI_ADAPTER_CONTAINER}
