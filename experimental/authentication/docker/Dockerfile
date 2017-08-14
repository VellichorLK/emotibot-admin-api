FROM golang:1.8-alpine AS build-env

ENV GOPREFIX /go/src/emotibot.com/emotigo
ARG PROJECT
RUN apk add --no-cache git

RUN mkdir -p ${GOPREFIX}
ADD . ${GOPREFIX}/
RUN cd ${GOPREFIX}/${PROJECT} && go-wrapper download && go-wrapper install



FROM alpine:3.5

ENV GOPREFIX /go/src/emotibot.com
COPY --from=build-env /go/bin/authentication /usr/local/bin/authentication

WORKDIR /usr/local/bin

CMD ["/usr/local/bin/authentication"]
