FROM golang:1.8-alpine AS buildEnv
ARG Module
ADD . /go/src/
RUN cd /go/src/${Module} &&go build -o transfer

FROM alpine:3.5
WORKDIR /usr/local/bin
RUN apk add --no-cache tzdata
ARG Module
RUN mkdir /usr/local/bin/tmp
COPY --from=buildEnv /go/src/${Module}/transfer /usr/local/bin/transfer
COPY ${Module}/docker/entrypoint.sh /usr/local/bin/entrypoint.sh
ENTRYPOINT [ "/usr/local/bin/entrypoint.sh" ]