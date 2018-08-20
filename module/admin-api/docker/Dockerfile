FROM docker-reg.emotibot.com.cn:55688/go-build:650dbada-alpine AS build-env

ENV GOPREFIX /go/src/emotibot.com/emotigo
ARG PROJECT

COPY ./module/${PROJECT} /go/src/emotibot.com/emotigo/module/${PROJECT}
COPY ./pkg /go/src/emotibot.com/emotigo/pkg
WORKDIR /go/src/emotibot.com/emotigo/module/${PROJECT}
RUN go test ./...
RUN go get -v && go install
RUN cd tool && for tool in `find . -type d -maxdepth 1 -mindepth 1`; do cd $tool && go install && cd ../; done;
RUN cd tool && for tool in `find . -type d -maxdepth 1 -mindepth 1`; do cd $tool && CGO_ENABLED=0 GOOS=linux go build -o ${tool}_linux && cp ${tool}_linux /go/bin/ && cd ../; done;


FROM alpine:3.5

ARG PROJECT
ENV GOPREFIX /go/src/emotibot.com
RUN apk add --no-cache tzdata

ENV SELF_RESOURCES http://docker-build01.emotibot.com:8000/selfLearnResources.tar.gz
RUN wget ${SELF_RESOURCES} -O /usr/local/share/resouc.tar.gz 
RUN cd /usr/local/share && tar -zxvf resouc.tar.gz && rm resouc.tar.gz

COPY --from=build-env /go/bin /usr/local/bin

WORKDIR /usr/local/bin
COPY ${PROJECT}/docker/entrypoint.sh entrypoint.sh
COPY ${PROJECT}/docker/env.template env.template
COPY ${PROJECT}/docker/files_init.sh files_init.sh
COPY ${PROJECT}/InitFiles InitFiles

CMD ["./entrypoint.sh", "env.template"]
