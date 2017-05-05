FROM golang:1.8-alpine AS build-env

ENV GOPREFIX /go/src/emotibot.com

RUN mkdir -p ${GOPREFIX}
ADD . ${GOPREFIX}
ARG PROJECT
RUN cd ${GOPREFIX}/${PROJECT} && go install


FROM alpine:3.5

COPY --from=build-env /go/bin/proxy /usr/local/bin/proxy

WORKDIR /usr/local/bin

CMD ["/usr/local/bin/proxy"]


