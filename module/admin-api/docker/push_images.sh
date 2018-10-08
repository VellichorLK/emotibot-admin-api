#!/bin/bash
source `pwd`/image_tags.sh;
docker push $IMAGE_NAME && docker rmi $IMAGE_NAME;
