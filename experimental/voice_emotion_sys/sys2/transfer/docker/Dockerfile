FROM golang:1.8-alpine AS build-env

COPY . /go/src
ARG Module
RUN cd /go/src/${Module} && go build -o /go/bin/transfer


FROM alpine:3.5
ARG Module
RUN mkdir -p /usr/src/app
COPY --from=build-env /go/bin/transfer /usr/src/app/transfer
COPY ${Module}/docker/entrypoint.sh /usr/src/app/entrypoint.sh
WORKDIR /usr/src/app

ENTRYPOINT ["/usr/src/app/entrypoint.sh"]
