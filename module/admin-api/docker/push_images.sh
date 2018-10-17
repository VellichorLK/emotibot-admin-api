#!/bin/bash
source `pwd`/image_tags.sh;
docker push $IMAGE_NAME && docker rmi $IMAGE_NAME;
docker push $BUILD_IMAGE_NAME && docker rmi $BUILD_IMAGE_NAME;