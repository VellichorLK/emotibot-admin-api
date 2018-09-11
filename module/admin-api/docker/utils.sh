#!/bin/sh

# Validate the Compose file to ensure no environment is missing
function valid_config() {
    if [ -z "$1" ]; then
        echo "Please specify the path of compose file."
        return 1
    fi

    cmd="docker-compose -f $1 config 2>&1 | grep 'Defaulting to a blank string'"

    if eval $cmd; then
        exit 1
    fi
}
