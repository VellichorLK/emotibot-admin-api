FROM golang:1.9-alpine AS build-env

COPY . /go/src
ARG Module
RUN cd /go/src/${Module} && go build -o /go/bin/tester


FROM alpine:3.5
RUN mkdir -p /usr/src/app/testfile
COPY --from=build-env /go/bin/tester /usr/src/app/tester
ARG Module
COPY --from=build-env /go/src/${Module}/entrypoint.sh /usr/src/app/entrypoint.sh

WORKDIR /usr/src/app
