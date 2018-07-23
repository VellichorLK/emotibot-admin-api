#!/bin/bash

ORIGIN_ENV=$1

if [ $# -ne 1 ]; then
    printf "add prefix for local env to use in deployment"
    printf "usage:\n"
    printf "\tadd_prefix.sh ../test.env\n"
    exit 1
fi

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

sed -e 's/\(^[a-zA-Z]\)/ADMIN_\1/' $DIR/$1