FROM golang:1.8-alpine AS build-env

ENV GOPREFIX /go/src/emotibot.com

RUN mkdir -p ${GOPREFIX}
COPY emotigo ${GOPREFIX}/emotigo
ARG PROJECT
RUN cd ${GOPREFIX}/${PROJECT} && go install


FROM alpine:3.5

ENV GOPREFIX /go/src/emotibot.com
COPY --from=build-env /go/bin/authentication /usr/local/bin/authentication

WORKDIR /usr/local/bin

CMD ["/usr/local/bin/authentication"]


