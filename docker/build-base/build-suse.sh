REPO=docker-reg.emotibot.com.cn:55688
CONTAINER=base/sles12sp3-golang1.9
TAG=20180416
DOCKER_IMAGE=$REPO/$CONTAINER:$TAG

cmd="docker build \
        -t $DOCKER_IMAGE \
        -f Dockerfile-SUSE ."
echo $cmd
eval $cmd
