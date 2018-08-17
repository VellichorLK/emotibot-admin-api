FROM docker-reg.emotibot.com.cn:55688/go-build:650dbada-alpine AS build-env

COPY ./ /go/src/emotibot.com/emotigo/module/openapi-adapter
RUN cd /go/src/emotibot.com/emotigo/module/openapi-adapter && go install

FROM alpine:3.5

COPY --from=build-env /go/bin /usr/local/bin

ENTRYPOINT ["openapi-adapter"]