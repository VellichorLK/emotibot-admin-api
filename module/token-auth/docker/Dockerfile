FROM docker-reg.emotibot.com.cn:55688/go-build:20180412_9359f34 AS build-env

ENV GOPREFIX /go/src/emotibot.com/emotigo
ARG PROJECT

COPY . /go/src/emotibot.com/emotigo/module
WORKDIR /go/src/emotibot.com/emotigo/module/${PROJECT}
RUN go-wrapper download && go-wrapper install
RUN cd tool && for tool in `find . -type d -maxdepth 1 -mindepth 1`; do cd $tool && go-wrapper install && cd ../; done;

FROM alpine:3.5

ARG PROJECT
ENV GOPREFIX /go/src/emotibot.com
RUN apk add --no-cache tzdata
COPY --from=build-env /go/bin /usr/local/bin

WORKDIR /usr/local/bin
CMD ["./token-auth"]
