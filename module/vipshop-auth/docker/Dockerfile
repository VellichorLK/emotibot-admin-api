FROM docker-reg.emotibot.com.cn:55688/vipshop-auth-adapter-base:20190327-368d1e62 AS build-env

ARG PROJECT

COPY . ${MODULEPREFIX}
RUN cd ${MODULEPREFIX}/${PROJECT} && go-wrapper install

FROM alpine:3.5

RUN echo http://mirrors.ustc.edu.cn/alpine/v3.7/main > /etc/apk/repositories; \
echo http://mirrors.ustc.edu.cn/alpine/v3.7/community >> /etc/apk/repositories
RUN apk add --no-cache tzdata

ARG PROJECT
COPY --from=build-env /go/bin/vipshop-auth /usr/local/bin/vipshop-auth

WORKDIR /usr/local/bin
COPY ${PROJECT}/docker/entrypoint.sh entrypoint.sh
COPY ${PROJECT}/docker/env.template env.template

CMD ["./entrypoint.sh", "env.template"]
