FROM golang:1.9.1-alpine AS build-env

COPY . /go/src
ARG Module
RUN cd /go/src/${Module} && CGO_ENABLED=0 go build -o /go/bin/web_service


FROM ubuntu:16.04
RUN apt-get update
RUN apt-get -y install glusterfs-client
RUN mkdir -p /usr/src/app/upload_file
COPY --from=build-env /go/bin/web_service /usr/src/app/web_service
ARG Module
COPY ${Module}/docker/entrypoint.sh /usr/src/app/entrypoint.sh
WORKDIR /usr/src/app

ENTRYPOINT ["/usr/src/app/entrypoint.sh"]
