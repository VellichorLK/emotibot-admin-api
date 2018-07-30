# Statistic

Emotibot statistic module based on Elasticsearch.

## Quick Start:

```
# Run static server on localhost machine:
$ go run *.go test.env

# Build docker image
$ ./docker/build.sh

# Run docker image
$ ./docker/run.sh ./docker/test.env
```

## Environment Variables:

```
STATISTIC_SERVER_PORT=8182                      // Server port
STATISTIC_SERVER_MYSQL_URL=172.17.0.1           // MySQL url
STATISTIC_SERVER_MYSQL_USER=root                // MySQL user name
STATISTIC_SERVER_MYSQL_PASS=password            // MySQL user password
STATISTIC_SERVER_ELASTICSEARCH_HOST=172.17.0.1  // ElasticSearch host url
STATISTIC_SERVER_ELASTICSEARCH_PORT=9200        // ElasticSearch host port
```

## ElasticSearch Query DSL Examples:
Please see ./examples
