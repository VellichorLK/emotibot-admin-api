#!/bin/bash

mount -t glusterfs glusterfscluster:/glusterfsvol /usr/src/app/upload_file

/usr/src/app/web_service