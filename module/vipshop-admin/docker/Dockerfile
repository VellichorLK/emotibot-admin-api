FROM docker-reg.emotibot.com.cn:55688/go-build:20171221_2d62f2b AS build-env

ENV GOPREFIX /go/src/emotibot.com/emotigo
ARG PROJECT

COPY . /go/src/emotibot.com/emotigo/module
RUN cd /go/src/emotibot.com/emotigo/module/${PROJECT} && go-wrapper download && go-wrapper install

FROM alpine:3.5

ARG PROJECT
ENV GOPREFIX /go/src/emotibot.com
RUN apk add --no-cache tzdata
COPY --from=build-env /go/bin/vipshop-admin /usr/local/bin/vipshop-admin

WORKDIR /usr/local/bin
COPY ${PROJECT}/docker/entrypoint.sh entrypoint.sh
COPY ${PROJECT}/docker/env.template env.template
COPY ${PROJECT}/docker/files_init.sh files_init.sh
COPY ${PROJECT}/docker/profile_rebuild.sh profile_rebuild.sh
COPY ${PROJECT}/InitFiles InitFiles

CMD ["./entrypoint.sh", "env.template"]
