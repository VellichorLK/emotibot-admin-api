#!/bin/bash
# ref: https://github.com/grafana/grafana-docker
CONTAINER="grafana"
TAG="4.1.1"
PORT=8088

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
if [ ! -d "$DIR/grafana_storage" ]; then
    mkdir -p "$DIR/grafana_storage"
fi

docker rm -f $CONTAINER
docker run -d --name=$CONTAINER -p $PORT:$PORT \
  -v "$DIR/grafana_storage":"/var/lib/grafana" \
  -v "$DIR/grafana.ini":"/etc/grafana/grafana.ini" \
  -e "GF_INSTALL_PLUGINS=grafana-piechart-panel,briangann-gauge-panel,btplc-trend-box-panel,briangann-datatable-panel" \
  $CONTAINER/$CONTAINER:$TAG
