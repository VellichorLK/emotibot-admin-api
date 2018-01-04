FROM golang:1.8-alpine AS build-env

ARG Module=vipshop-fake-cas

COPY ./ /go/src/${Module}/
RUN cd /go/src/${Module} && go-wrapper download  && CGO_ENABLED=0 go build -o /go/bin/${Module}

FROM alpine:3.5
ARG Module=vipshop-fake-cas
COPY --from=build-env /go/bin/${Module} /usr/bin/
ADD ./dist /usr/bin/dist/
RUN mv /usr/bin/${Module} /usr/bin/goExecutable
WORKDIR /usr/bin
ENTRYPOINT [ "/usr/bin/goExecutable" ]
