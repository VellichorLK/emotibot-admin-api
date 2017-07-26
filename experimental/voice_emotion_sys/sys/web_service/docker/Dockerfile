FROM golang:1.8-alpine AS build-env

COPY . /go/src
ARG Module
RUN cd /go/src/${Module} && go build -o /go/bin/web_service


FROM alpine:3.5
RUN mkdir -p /usr/src/app/upload_file
COPY --from=build-env /go/bin/web_service /usr/src/app/web_service

WORKDIR /usr/src/app

CMD ["/usr/src/app/web_service"]
