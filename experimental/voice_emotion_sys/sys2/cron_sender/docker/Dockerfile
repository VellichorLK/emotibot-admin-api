FROM golang:1.8-alpine AS build-env

COPY . /go/src
ARG Module
RUN cd /go/src/${Module} && go build -o /go/bin/${Module}


FROM alpine:3.5
ARG Module
RUN mkdir -p /usr/src/app
COPY ./${Module}/template /usr/src/app/template
COPY --from=build-env /go/bin/${Module} /usr/src/app/cron_sender


WORKDIR /usr/src/app

CMD ["/usr/src/app/cron_sender"]
