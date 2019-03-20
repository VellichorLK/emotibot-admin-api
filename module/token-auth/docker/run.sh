#!/bin/sh

# Exit immediately if a command exits with a non-zero status
set -e

DIR="$( cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [ -z "$1" ]; then
    ENV_FILE="${DIR}/run.env"
    echo "Using default environment file: ${ENV_FILE}"
else
    ENV_FILE=$1
fi

source ${DIR}/utils.sh

# Get docker image tags and export to environment variables
set -a
source ${DIR}/image_tags.sh

# Export environment variables from ${ENV_FILE}
ENV_PATH="${DIR}/${ENV_FILE}"

while read line
do
    eval ${line}
done < ${ENV_FILE}

valid_config "${DIR}/docker-compose.yml"

# Run docker image
docker-compose -f ${DIR}/docker-compose.yml up -d
