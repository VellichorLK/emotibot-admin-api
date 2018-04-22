# Mysql Network Database Deployment

*This is a testing envoirment setup for now, DO NOT USE IN PRODUTION*

A single host docker-compose is used, potentially swarm mode but need to address the ip problem.

If any of its container got error at the start, you may have to restart all container. This script have not mount the data volume out yet, so it could lost all data if the all data container got killed.


## How to start

`docker-compose up -d`

## How to remove

`docker-compose down`

## Components

* 1 management node
* 2 Data node
* 2 sql node