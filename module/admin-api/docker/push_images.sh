#!/bin/bash
source `pwd`/image_tags.sh;
docker push $IMAGE_NAME;
