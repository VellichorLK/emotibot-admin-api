FROM harbor.emotibot.com/library/go-build:80ac6858-1.10-standard AS build-env

ENV GOPREFIX /go/src/emotibot.com/emotigo
ARG PROJECT

COPY ./module/${PROJECT} /go/src/emotibot.com/emotigo/module/${PROJECT}
COPY ./pkg /go/src/emotibot.com/emotigo/pkg
WORKDIR /go/src/emotibot.com/emotigo/module/${PROJECT}
RUN go get -v && go install -v
RUN cd tool && for tool in `find . -type d -maxdepth 1 -mindepth 1`; do cd $tool && go install -v && cd ../; done;

FROM centos:centos7.4.1708

ARG PROJECT
ENV GOPREFIX /go/src/emotibot.com
RUN yum install --disableplugin=fastestmirror -y tzdata
COPY --from=build-env /go/bin /usr/local/bin

WORKDIR /usr/local/bin
CMD ["./token-auth"]
