#!/bin/sh

GIT_HEAD="$(git rev-parse --short=7 HEAD)"
GIT_DATE=$(git log HEAD -n1 --pretty='format:%cd' --date=format:'%Y%m%d-%H%M')

# Use docker-reg for now, it will change to harbor after harbor usage is confirmed
REPO="harbor.emotibot.com"
PROJECT="bfop"
CONTAINER="token-auth"
TAG="${GIT_HEAD}-${GIT_DATE}"

BUILD_IMAGE_NAME="${REPO}/${PROJECT}/${CONTAINER}-build:${TAG}"
IMAGE_NAME="${REPO}/${PROJECT}/${CONTAINER}:${TAG}"
CONTAINER_NAME=${CONTAINER}
