#!/bin/bash

GIT_REPO_DIR="vendor_deployment_scripts"
GITLAB_URL="ssh://git@gitlab.emotibot.com:10022"
CLUSTERS="10.0.0.46 10.0.0.47 10.0.0.48 10.0.0.50 10.0.0.51 10.0.0.52 10.0.0.53 10.0.0.54 10.0.0.55 10.0.0.56 10.0.0.57 10.0.0.58 10.0.0.59"
DOCKER_IMG="docker-reg.emotibot.com.cn:55688/filebeat:5.1.2"

echo "pull image $DOCKER_IMG from SH"
./pull_image_from_SH.sh $DOCKER_IMG

for cluster in $CLUSTERS 
do 
    ssh deployer@$cluster "
        pwd
        if [ ! -d $GIT_REPO_DIR ]; then
            git clone $GITLAB_URL/deployment/$GIT_REPO_DIR.git
            cd $GIT_REPO_DIR/filebeat
        else 
            cd $GIT_REPO_DIR
            git pull
            cd filebeat 
        fi
        ./restart.sh
        docker ps | grep filebeat | grep -v grep
    "
done
