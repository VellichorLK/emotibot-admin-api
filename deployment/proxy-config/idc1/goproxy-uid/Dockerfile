FROM golang:1.8-alpine AS build-env

ENV GOPREFIX /go/src/emotibot.com

RUN mkdir -p ${GOPREFIX}
ADD emotigo ${GOPREFIX}/emotigo
ARG PROJECT
RUN cd ${GOPREFIX}/${PROJECT} && go install
WORKDIR ${GOPREFIX}/${PROJECT}
RUN cd fakeserver && go build -o fakeserver .
RUN cp fakeserver/fakeserver /go/bin


FROM alpine:3.5

COPY --from=build-env /go/bin/goproxy-uid /usr/local/bin/goproxy-uid
RUN mkdir -p /app/fakeserver
COPY --from=build-env /go/bin/fakeserver /app/fakeserver/fakeserver
WORKDIR /usr/local/bin

CMD ["/usr/local/bin/goproxy-uid"]


