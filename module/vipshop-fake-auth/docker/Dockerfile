FROM golang:1.8-alpine AS build-env


ARG PROJECT
RUN apk add --no-cache git

COPY . /go/src/emotibot.com/emotigo/module
RUN cd /go/src/emotibot.com/emotigo/module/${PROJECT} && go-wrapper download 
RUN cd /go/src/emotibot.com/emotigo/module/${PROJECT} && go-wrapper install

FROM alpine:3.5

ARG PROJECT
ENV GOPREFIX /go/src/emotibot.com
COPY --from=build-env /go/bin/vipshop-fake-auth /usr/local/bin/vipshop-fake-auth

WORKDIR /usr/local/bin
COPY ${PROJECT}/docker/entrypoint.sh entrypoint.sh

CMD ["./entrypoint.sh", "env.template"]
