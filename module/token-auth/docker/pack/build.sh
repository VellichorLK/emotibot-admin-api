#!/bin/sh

# Exit immediately if a command exits with a non-zero status
set -e

DIR="$( cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/../utils.sh

# Get docker image tags and export to environment variables
set -a
source ${DIR}/../image_tags.sh

MODULE_PATH="$( cd "$( dirname "${DIR}/../../.." )" && pwd)"
MODULE="$( basename ${MODULE_PATH} )"

BUILD_CONTEXT="${DIR}/../../../.."
DOCKER_FILE="${DIR}/Dockerfile.pack"

valid_config "${DIR}/docker-compose.yml"

# Build docker image
# 透过传入 BUILD_IMAGE_NAME build arg 来指定在 build stage 所生成的 docker image name
docker-compose -f ${DIR}/docker-compose.yml build --build-arg=BUILD_IMAGE_NAME=${BUILD_IMAGE_NAME} --build-arg=PROJECT=${MODULE}
