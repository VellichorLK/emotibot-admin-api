FROM golang:1.8-alpine AS build-env

COPY . /go/src
ARG Module
RUN cd /go/src/${Module} && go build -o /go/bin/render


FROM alpine:3.5
RUN apk update && apk add ca-certificates
RUN mkdir -p /usr/src/app
COPY --from=build-env /go/bin/render /usr/src/app/render

WORKDIR /usr/src/app

CMD ["/usr/src/app/render"]
