#!/bin/sh

# Exit immediately if a command exits with a non-zero status
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Get docker image tags and export to environment variables
set -a
source ${DIR}/image_tags.sh

# Start building docker image
${DIR}/build/build.sh

# Start to pack
${DIR}/pack/build.sh
