#!/bin/sh

GIT_HEAD="$(git rev-parse --short HEAD)"
GIT_DATE=$(git log HEAD -n1 --pretty='format:%cd' --date=format:'%Y%m%d-%H%M')

# Use docker-reg for now, it will change to harbor after harbor usage is confirmed
# ADMIN_REPO="harbor.emotibot.com/bfop"
ADMIN_REPO="docker-reg.emotibot.com.cn:55688"
ADMIN_CONTAINER="admin-api"
ADMIN_TAG="${GIT_HEAD}_${GIT_DATE}"

ADMIN_BUILD_IMAGE_NAME="${ADMIN_REPO}/${ADMIN_CONTAINER}_build:${ADMIN_TAG}"
ADMIN_IMAGE_NAME="${ADMIN_REPO}/${ADMIN_CONTAINER}:${ADMIN_TAG}"
ADMIN_CONTAINER_NAME=${ADMIN_CONTAINER}
