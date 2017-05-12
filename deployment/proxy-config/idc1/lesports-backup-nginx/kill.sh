#!/bin/bash


docker rm -fv $(docker ps --format={{.Names}} | grep lesports-backup-nginx)
