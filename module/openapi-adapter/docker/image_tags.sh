#!/bin/sh

GIT_HEAD="$(git rev-parse --short=7 HEAD)"
GIT_DATE=$(git log HEAD -n1 --pretty='format:%cd' --date=format:'%Y%m%d-%H%M')

# Use docker-reg for now, it will change to harbor after harbor usage is confirmed
# REPO="harbor.emotibot.com/bfop"
REPO="docker-reg.emotibot.com.cn:55688"
CONTAINER="openapi-adapter"
TAG="${GIT_HEAD}-${GIT_DATE}"

BUILD_IMAGE_NAME="${REPO}/${CONTAINER}-build:${TAG}"
IMAGE_NAME="${REPO}/${CONTAINER}:${TAG}"
CONTAINER_NAME=${CONTAINER}
