#!/bin/bash

REPO=docker-reg.emotibot.com.cn:55688
# The name of the container, should use the name of the repo is possible
# <EDIT_ME>
CONTAINER=golang-dev
# </EDIT_ME>
TAG=latest
DOCKER_IMAGE=$REPO/$CONTAINER:$TAG


# Note that the user id begins from 501 for MacOS but from 1,000 for linux,
# the work around is to set the uid as 1,000 directly for docker user.
# Ref: https://en.wikipedia.org/wiki/User_identifier
os=$(uname)
if [[ "$os" == 'Darwin' ]]; then
  USER_ID=1000
  GROUP_ID=1000
else
  USER_ID=$(id -u)
  GROUP_ID=$(id -g)
fi

# TTY (i.e., -it) is only enabled when no additional command (to execute) is appended.
if [ "$#" -eq 0 ]; then
  DOCKER_OPT="-it --rm"
else
  DOCKER_OPT="--rm"
fi


# Mount the $PROJROOT into docker
ROOT=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
PROJROOT=$ROOT/../../../emotigo

# Start docker
docker rm -f -v $CONTAINER
cmd="docker run $DOCKER_OPT \
    --net=host \
    -e "USERID=$USER_ID"    `# Export user id`\
    -e "GROUPID=$GROUP_ID"  `# Export group id`\
    -e "PYTHONUNBUFFERED=1" `# Force stdio, stdout, and stderr unbuffered`\
    -v $ROOT/cache/src:/go/src \
    -v $ROOT/cache/bin:/go/bin \
    -v $PROJROOT:/go/src/emotibot.com/emotigo \
    $DOCKER_IMAGE $@"       `# Run additional command`
echo $cmd
$cmd
