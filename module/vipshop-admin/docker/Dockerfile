FROM docker-reg.emotibot.com.cn:55688/go-build:20180319_d6de838 AS build-env

ENV GOPREFIX /go/src/emotibot.com/emotigo
ARG PROJECT

COPY . /go/src/emotibot.com/emotigo/module
RUN cd /go/src/emotibot.com/emotigo/module/${PROJECT} && go-wrapper download && go-wrapper install

FROM alpine:3.5

ENV SELF_RESOURCES http://192.168.3.251:8181/selfLearnResources.tar.gz
RUN wget ${SELF_RESOURCES} -O /usr/local/share/resouc.tar.gz 
RUN cd /usr/local/share && tar -zxvf resouc.tar.gz && rm resouc.tar.gz

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
