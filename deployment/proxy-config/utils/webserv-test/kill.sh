#!/bin/bash


docker rm -fv $(docker ps --format={{.Names}} | grep nginx-lesports-backup)
