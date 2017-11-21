FROM golang:1.8-alpine AS build-env

ARG PROJECT
RUN apk add --no-cache git
ENV MODULEPREFIX /go/src/emotibot.com/emotigo/module
COPY . ${MODULEPREFIX}
RUN cd ${MODULEPREFIX}/${PROJECT} && go-wrapper download && go-wrapper install

FROM alpine:3.5

ARG PROJECT
COPY --from=build-env /go/bin/vipshop-auth /usr/local/bin/vipshop-auth

WORKDIR /usr/local/bin
COPY ${PROJECT}/docker/entrypoint.sh entrypoint.sh
COPY ${PROJECT}/docker/env.template env.template

CMD ["./entrypoint.sh", "env.template"]
